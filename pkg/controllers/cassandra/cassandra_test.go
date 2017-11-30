package cassandra_test

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	informers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	kubeinformers "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions"

	navigatorfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
)

type fixture struct {
	t                *testing.T
	kclient          *fake.Clientset
	podWatch         *watch.FakeWatcher
	statefulSetWatch *watch.FakeWatcher
	nclient          *navigatorfake.Clientset
	cassWatch        *watch.FakeWatcher
	recorder         *record.FakeRecorder
	syncSuccess      chan struct{}
	finished         chan struct{}
}

func NewFixture(t *testing.T) *fixture {
	nclient := navigatorfake.NewSimpleClientset()
	cassWatch := watch.NewFake()
	nclient.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(cassWatch, nil),
	)
	kclient := fake.NewSimpleClientset()
	podWatch := watch.NewFake()
	kclient.PrependWatchReactor(
		"pods",
		clienttesting.DefaultWatchReactor(podWatch, nil),
	)
	statefulSetWatch := watch.NewFake()
	kclient.PrependWatchReactor(
		"statefulsets",
		clienttesting.DefaultWatchReactor(statefulSetWatch, nil),
	)

	return &fixture{
		t:                t,
		kclient:          kclient,
		podWatch:         podWatch,
		statefulSetWatch: statefulSetWatch,
		nclient:          nclient,
		cassWatch:        cassWatch,
		recorder:         record.NewFakeRecorder(0),
		syncSuccess:      make(chan struct{}),
		finished:         make(chan struct{}),
	}
}

func (f *fixture) run() {
	go func() {
		for e := range f.recorder.Events {
			f.t.Logf("EVENT: %q", e)
			if strings.Contains(e, cassandra.SuccessSync) {
				f.syncSuccess <- struct{}{}
			}
		}
		close(f.syncSuccess)
		close(f.finished)
	}()
}

// NewCassandra sets up event handlers for the supplied informers.
func TestCassandraControllerIntegration(t *testing.T) {
	f := NewFixture(t)
	ctx := &controllers.Context{
		Client:                    f.kclient,
		NavigatorClient:           f.nclient,
		Recorder:                  f.recorder,
		KubeSharedInformerFactory: kubeinformers.NewSharedInformerFactory(f.kclient, 0),
		SharedInformerFactory:     informers.NewSharedInformerFactory(f.nclient, 0),
		Namespace:                 "namespace-not-used-in-this-test",
	}
	controller := cassandra.CassandraControllerFromContext(ctx)

	stopCh := make(chan struct{})
	ctx.SharedInformerFactory.Start(stopCh)
	ctx.KubeSharedInformerFactory.Start(stopCh)
	controllerFinished := make(chan struct{})
	go func() {
		defer func() {
			close(f.recorder.Events)
			close(controllerFinished)
		}()
		err := controller.Run(1, stopCh)
		if err != nil {
			t.Error(err)
		}
	}()
	f.run()
	defer func() {
		close(stopCh)
		<-controllerFinished
		<-f.finished
	}()

	cluster := casstesting.ClusterForTest()
	f.cassWatch.Add(cluster)
	select {
	case <-f.syncSuccess:
	case <-time.After(time.Second):
		t.Error("Timeout waiting for controller sync")
	}
	ss := &apps.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-set",
			Namespace: cluster.Namespace,
			UID:       types.UID("test"),
		},
	}
	ss.SetOwnerReferences(
		append(
			ss.GetOwnerReferences(),
			*metav1.NewControllerRef(
				cluster,
				cluster.GetObjectKind().GroupVersionKind(),
			),
		),
	)
	f.statefulSetWatch.Add(ss)

	pod := &v1.Pod{}
	pod.SetName("some-pod")
	pod.SetNamespace(cluster.Namespace)
	pod.SetOwnerReferences(
		append(
			pod.GetOwnerReferences(),
			*metav1.NewControllerRef(
				ss,
				ss.GetObjectKind().GroupVersionKind(),
			),
		),
	)
	f.podWatch.Add(pod)
	select {
	case <-f.syncSuccess:
	case <-time.After(time.Second * 5):
		t.Error("Timeout waiting for controller sync")
	}
}
