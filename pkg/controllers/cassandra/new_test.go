package cassandra_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/glog"
	kubeinformers "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corelisters "k8s.io/client-go/listers/core/v1"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type fix struct {
	t       *testing.T
	kclient *fake.Clientset
	i       kubeinformers.SharedInformerFactory
	w       *watch.FakeWatcher
}

func NewFix(t *testing.T) *fix {
	kclientWrite := fake.NewSimpleClientset()
	kclientRead := fake.NewSimpleClientset()

	i := kubeinformers.NewSharedInformerFactory(kclientRead, 0)
	fix := &fix{
		t:       t,
		kclient: kclientWrite,
		i:       i,
		w:       watch.NewFake(),
	}
	kclientRead.PrependWatchReactor(
		"*",
		clienttesting.DefaultWatchReactor(
			fix.w,
			nil,
		),
	)
	kclientWrite.PrependReactor(
		"create",
		"*",
		fix.react,
	)
	return fix
}

func (f *fix) react(a clienttesting.Action) (bool, runtime.Object, error) {
	var o runtime.Object
	switch action := a.(type) {
	case clienttesting.CreateAction:
		o = action.GetObject()
		f.w.Add(o)
	default:
		f.t.Errorf("Unexpected action: %#v", action)
	}
	return true, o, nil
}

func (f *fix) Run(stopCh chan struct{}) {
	f.i.Start(stopCh)
	_ = f.i.WaitForCacheSync(stopCh)
}

type Controller struct {
	client kubernetes.Interface
	lister corelisters.PodLister
	queue  workqueue.RateLimitingInterface
}

func NewController(
	client kubernetes.Interface,
	pods coreinformers.PodInformer,
	rcs coreinformers.ReplicationControllerInformer,
) *Controller {

	queue := workqueue.NewRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
	)

	c := &Controller{
		client: client,
		lister: pods.Lister(),
		queue:  queue,
	}

	rcs.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Error(err)
					return
				}
				queue.Add(key)
			},
		},
	)

	pods.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				_, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Error(err)
					return
				}
				pod, ok := obj.(*v1.Pod)
				if !ok {
					glog.Error("not a pod")
					return
				}
				c.handlePod(pod)
			},
		},
	)
	return c
}

func (c *Controller) handlePod(pod *v1.Pod) {
	glog.V(4).Infof("handlePod: %#v", pod)
	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef == nil {
		glog.V(4).Infof("ignoring pod without owner")
		return
	}
	if ownerRef.Kind != "ReplicationController" {
		glog.V(4).Infof("ignoring pod because it is not controlled by an RC")
		return
	}
	c.queue.Add(fmt.Sprintf("%s/%s", pod.Namespace, ownerRef.Name))
}

func TestFakeClient(t *testing.T) {
	f := NewFix(t)
	NewController(
		f.kclient,
		f.i.Core().V1().Pods(),
		f.i.Core().V1().ReplicationControllers(),
	)

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-pod",
			Namespace: v1.NamespaceDefault,
			UID:       types.UID("test"),
		},
	}
	stopCh := make(chan struct{})
	f.Run(stopCh)
	_, err := f.kclient.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(time.Second)
}
