package pilotcontroller

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigatorinformers "github.com/jetstack/navigator/pkg/client/informers/externalversions/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
)

const (
	SuccessSync        = "SuccessSync"
	MessageSuccessSync = "Successfully synced pods"
)

type Interface interface {
	Run(workers int, stopCh <-chan struct{}) error
}

type pilotController struct {
	state              *controllers.State
	pilotsListerSynced cache.InformerSynced
	podsListerSynced   cache.InformerSynced
	queue              workqueue.RateLimitingInterface
	recorder           record.EventRecorder
}

var _ Interface = &pilotController{}

func New(
	pilots navigatorinformers.PilotInformer,
	pods coreinformers.PodInformer,
	state *controllers.State,
	recorder record.EventRecorder,
) Interface {
	queue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"pilot",
	)
	pods.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					_, found := t.Labels[v1alpha1.PilotLabel]
					if found {
						glog.V(4).Infof("Adding piloted pod to queue: %s/%s", t.Namespace, t.Name)
					}
					return found
				default:
					utilruntime.HandleError(
						fmt.Errorf("object not expected: %T", obj),
					)
					return false
				}
			},
			Handler: &controllers.QueuingEventHandler{Queue: queue},
		},
	)
	pilots.Informer().AddEventHandler(&controllers.QueuingEventHandler{Queue: queue})
	return &pilotController{
		state:              state,
		pilotsListerSynced: pilots.Informer().HasSynced,
		podsListerSynced:   pods.Informer().HasSynced,
		queue:              queue,
		recorder:           recorder,
	}
}

func NewFromContext(ctx *controllers.Context) Interface {
	return New(
		ctx.SharedInformerFactory.Navigator().V1alpha1().Pilots(),
		ctx.KubeSharedInformerFactory.Core().V1().Pods(),
		controllers.StateFromContext(ctx),
		ctx.Recorder,
	)
}

func init() {
	controllers.Register(
		"Pilot",
		func(ctx *controllers.Context) controllers.Interface {
			return NewFromContext(ctx).Run
		},
	)
}

// Run is the main event loop
func (c *pilotController) Run(workers int, stopCh <-chan struct{}) error {
	glog.Infof("Starting Pilot controller")
	if !cache.WaitForCacheSync(
		stopCh,
		c.pilotsListerSynced,
		c.podsListerSynced,
	) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.Until(c.worker, time.Second, stopCh)
		}()
	}

	<-stopCh
	c.queue.ShutDown()
	glog.V(4).Infof("Shutting down Pilot controller workers...")
	wg.Wait()
	glog.V(4).Infof("Pilot controller workers stopped.")
	return nil
}

func (c *pilotController) worker() {
	glog.V(4).Infof("start worker loop")
	for c.processNextWorkItem() {
		glog.V(4).Infof("processed work item")
	}
	glog.V(4).Infof("exiting worker loop")
}

func (c *pilotController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)
	glog.V(4).Infof("processing %#v", key)

	if err := c.sync(key.(string)); err != nil {
		glog.Infof(
			"Error syncing pod %v, requeuing: %v",
			key.(string), err,
		)
		c.queue.AddRateLimited(key)
	} else {
		c.queue.Forget(key)
	}
	return true
}

func (c *pilotController) createPilot(pod *v1.Pod) (*v1alpha1.Pilot, error) {
	pilot := PilotForPod(pod)
	pilot, err := c.state.NavigatorClientset.NavigatorV1alpha1().
		Pilots(pilot.Namespace).Create(pilot)
	if k8sErrors.IsAlreadyExists(err) {
		return pilot, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to create pilot")
	}
	glog.V(4).Infof("Created pilot %s/%s.", pilot.Namespace, pilot.Name)
	return pilot, nil
}

func (c *pilotController) sync(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		glog.Errorf("Unable to parse key: %#v", err)
		return nil
	}
	pod, err := c.state.PodLister.Pods(namespace).Get(name)
	podFound := err == nil
	if err != nil && !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unable to get pod")
	}
	if !podFound {
		glog.V(4).Infof("Pod not found: %s/%s", namespace, name)
		return nil
	}

	if pod.DeletionTimestamp != nil {
		glog.V(4).Infof("Pod is being deleted: %s/%s", namespace, name)
		return nil
	}

	pilot, err := c.state.PilotLister.Pilots(namespace).Get(name)
	pilotFound := err == nil
	if err != nil && !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unable to get pilot")
	}
	if pilotFound {
		glog.V(4).Infof("Pilot already exists found: %s/%s", namespace, name)
		return nil
	}
	glog.V(4).Infof("Creating pilot %s/%s.", namespace, name)
	pilot, err = c.createPilot(pod)
	if err == nil {
		c.recorder.Event(
			pilot,
			v1.EventTypeNormal,
			SuccessSync,
			MessageSuccessSync,
		)
	}
	return err
}

func PilotForPod(pod *v1.Pod) *v1alpha1.Pilot {
	return &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			// Pilot is owned by its pod,
			// and will be garbage collected at the same time as the pod.
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(
					pod,
					// https://github.com/kubernetes/kubernetes/issues/63676
					schema.GroupVersionKind{
						Group:   v1.SchemeGroupVersion.Group,
						Version: v1.SchemeGroupVersion.Version,
						Kind:    "Pod",
					},
				),
			},
		},
	}
}
