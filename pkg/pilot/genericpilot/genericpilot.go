package genericpilot

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/process"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/scheduler"
)

type GenericPilot struct {
	Options Options

	// TODO: remove use of the kubernetes clientset. Absorb required
	// functionality into the navigator api group
	kubeClientset kubernetes.Interface
	client        clientset.Interface

	pilotLister         listersv1alpha1.PilotLister
	pilotInformerSynced cache.InformerSynced

	queue    workqueue.RateLimitingInterface
	recorder record.EventRecorder

	// process is a reference to a process manager for the application this
	// Pilot manages
	process process.Interface
	// phase is the current phase of this Pilot. This is used as a source of
	// truth within Pilots, as we cannot rely on the pilot.status block being
	// up to date
	lastCompletedPhase v1alpha1.PilotPhase
	// shutdown is true when the process has been told to gracefully exit. This
	// is used to signal preStop hooks to run
	shutdown bool
	// lock is used internally to coordinate updates to fields on the
	// GenericPilot structure
	lock sync.Mutex
	// scheduledWorkQueue is used to periodically re-sync 'this' Pilot resource.
	scheduledWorkQueue scheduler.ScheduledWorkQueue
}

// only run one worker to prevent threading issues when dealing with processes
const workers = 1

func (g *GenericPilot) enqueuePilot(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	g.queue.AddRateLimited(key)
}

func (g *GenericPilot) WaitForCacheSync(stopCh <-chan struct{}) error {
	if !cache.WaitForCacheSync(stopCh, g.pilotInformerSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	return nil
}

func (g *GenericPilot) Run() error {
	glog.Infof("Starting generic pilot controller")

	// setup healthz handlers
	g.serveHealthz()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		wait.Until(g.worker, time.Second, g.Options.StopCh)
	}()

	<-g.Options.StopCh
	glog.V(4).Infof("Shutdown signal received")
	// set g.shutdown = true to signal preStop hooks to run
	g.shutdown = true
	glog.V(4).Infof("Waiting for process exit and hooks to execute")
	// wait until postStop hooks have run
	wait.Poll(time.Second*1, time.Minute*10, func() (bool, error) {
		return g.lastCompletedPhase == v1alpha1.PilotPhasePostStop, nil
	})
	glog.V(4).Infof("Shutting down workqueue")
	// shutdown the worker queue
	g.queue.ShutDown()
	glog.V(4).Infof("Shutting down generic pilot controller workers")
	// wait for workers to exit
	wg.Wait()
	glog.V(4).Infof("Generic pilot controller workers stopped")
	return nil
}
