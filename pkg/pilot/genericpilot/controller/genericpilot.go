package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	informersv1alpha1 "github.com/jetstack/navigator/pkg/client/informers/externalversions/navigator/v1alpha1"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/controller/scheduler"
)

type Controller struct {
	// SyncFunc is the implementor-supplied sync function
	syncFunc func(*v1alpha1.Pilot) error
	// pilotName is the name of 'this' Pilot
	// pilotNamespace is the namespace of 'this' Pilot
	pilotName, pilotNamespace string

	// TODO: remove use of the kubernetes clientset. Absorb required
	// functionality into the navigator api group
	kubeClientset kubernetes.Interface
	clientset     clientset.Interface

	pilotLister         listersv1alpha1.PilotLister
	pilotInformerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
	// scheduledWorkQueue is used to periodically re-sync 'this' Pilot resource.
	scheduledWorkQueue scheduler.ScheduledWorkQueue
	// lock is used internally to coordinate updates to fields on the
	// Controller structure
	lock sync.Mutex
	// cachedThisPilot is a reference to a Pilot resource for 'this' Pilot.
	// It may be out of date, and it should *never* be manipulated.
	// This is especially useful for circumstances where the Pilot resource is
	// no longer available, either due to a network outage, the namespace
	// containing the Pilot resource being deleted, or any other circumstance
	// leading to the Pilot lister to not contain a reference to 'this' pilot.
	cachedThisPilot *v1alpha1.Pilot
}

const (
	controllerAgentName = "generic-pilot"
)

type Options struct {
	PilotName, PilotNamespace string
	SyncFunc                  func(*v1alpha1.Pilot) error
	KubeClientset             kubernetes.Interface
	Clientset                 clientset.Interface
	PilotInformer             informersv1alpha1.PilotInformer
}

func NewController(opts Options) *Controller {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pilots")
	ctrl := &Controller{
		pilotName:           opts.PilotName,
		pilotNamespace:      opts.PilotNamespace,
		syncFunc:            opts.SyncFunc,
		kubeClientset:       opts.KubeClientset,
		clientset:           opts.Clientset,
		pilotLister:         opts.PilotInformer.Lister(),
		pilotInformerSynced: opts.PilotInformer.Informer().HasSynced,
		queue:               queue,
	}
	ctrl.scheduledWorkQueue = scheduler.NewScheduledWorkQueue(ctrl.enqueuePilot)

	opts.PilotInformer.Informer().AddEventHandler(&controllers.QueuingEventHandler{queue})

	return ctrl
}

// only run one worker to prevent threading issues when dealing with processes
const workers = 1

func (g *Controller) enqueuePilot(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	g.queue.AddRateLimited(key)
}

func (g *Controller) WaitForCacheSync(stopCh <-chan struct{}) error {
	if !cache.WaitForCacheSync(stopCh, g.pilotInformerSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	return nil
}

func (g *Controller) Run(stopCh <-chan struct{}) error {
	glog.Infof("Starting generic pilot controller")
	if err := g.WaitForCacheSync(stopCh); err != nil {
		return fmt.Errorf("error waiting for controller caches to sync: %s", err)
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			wait.Until(g.worker, time.Second, stopCh)
		}()
	}

	<-stopCh
	glog.V(4).Infof("Shutdown signal received. Shutting down workqueue...")
	g.queue.ShutDown()

	glog.V(4).Infof("Shutting down generic pilot controller workers")
	wg.Wait()

	glog.V(4).Infof("Generic pilot controller workers stopped")
	return nil
}
