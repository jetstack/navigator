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

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/process"
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

	process process.Interface
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
	g.queue.ShutDown()
	glog.V(4).Infof("Shutting down generic pilot controller workers...")
	wg.Wait()
	glog.V(4).Infof("Generic pilot controller workers stopped.")
	return nil
}
