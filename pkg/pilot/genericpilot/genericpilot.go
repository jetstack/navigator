package genericpilot

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/controller"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/processmanager"
)

type GenericPilot struct {
	Options Options

	// TODO: remove use of the kubernetes clientset. Absorb required
	// functionality into the navigator api group
	kubeClientset kubernetes.Interface
	client        clientset.Interface

	pilotLister listersv1alpha1.PilotLister

	recorder record.EventRecorder

	controller *controller.Controller
	// process is a reference to a process manager for the application this
	// Pilot manages
	process processmanager.Interface
	// shutdown is true when the process has been told to gracefully exit
	shutdown bool
	// lock is used internally to coordinate updates to fields on the
	// GenericPilot structure
	lock sync.Mutex
}

func (g *GenericPilot) Run() error {
	glog.Infof("Starting generic pilot controller")

	// setup healthz handlers
	g.serveHealthz()

	ctrlStopCh := make(chan struct{})
	defer close(ctrlStopCh)
	go g.controller.Run(ctrlStopCh)

	var err error
	// block until told to shutdown
	select {
	case <-g.Options.StopCh:
	case <-g.waitForProcess():
		if err = g.process.Error(); err != nil {
			glog.V(4).Infof("Underlying process failed")
		} else {
			glog.V(4).Infof("Underlying process unexpectedly exited")
		}
	}

	glog.V(4).Infof("Shutdown signal received")
	thisPilot, err := g.controller.ThisPilot()
	if err != nil {
		return err
	}
	return g.stop(thisPilot)
}

// waitForProcess will return a chan that will be closed once the underlying
// subprocess exits. This function exists to 'mask' the fact the process may
// not ever exist/be started (as starting the process relies on the Pilot
// resource existing in the API).
func (g *GenericPilot) waitForProcess() <-chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		for {
			if g.process != nil {
				break
			}
			time.Sleep(2)
		}
		<-g.process.Wait()
	}()
	return out
}
