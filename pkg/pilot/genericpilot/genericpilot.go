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
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/leaderelection"
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
	lock    sync.Mutex
	elector leaderelection.Interface
}

func (g *GenericPilot) Run() error {
	glog.Infof("Starting generic pilot controller")

	// setup healthz handlers
	g.serveHealthz()

	ctrlStopCh := make(chan struct{})
	defer close(ctrlStopCh)

	var err error
	// block until told to shutdown
	select {
	case <-g.Options.StopCh:
		glog.Infof("Shutdown signal received")
	case <-g.waitForProcess():
		if err = g.process.Error(); err != nil {
			glog.Errorf("Underlying process failed with error: %s", err)
		} else {
			glog.Errorf("Underlying process unexpectedly exited")
		}
	case err = <-g.runController(ctrlStopCh):
		if err != nil {
			glog.Errorf("Control loop failed with error: %s", err)
		} else {
			glog.Errorf("Control loop unexpectedly exited")
		}
	case err = <-g.runElector(ctrlStopCh):
		if err != nil {
			glog.Errorf("Leader elector failed with error: %s", err)
		} else {
			glog.Errorf("Leader elector unexpectedly exited")
		}
	}

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

func (g *GenericPilot) runController(stopCh <-chan struct{}) <-chan error {
	out := make(chan error, 1)
	go func() {
		defer close(out)
		out <- g.controller.Run(stopCh)
	}()
	return out
}

func (g *GenericPilot) runElector(stopCh <-chan struct{}) <-chan error {
	out := make(chan error, 1)
	go func() {
		defer close(out)
		out <- g.elector.Run()
	}()
	return out
}

func (g *GenericPilot) Elector() leaderelection.Interface {
	return g.elector
}
