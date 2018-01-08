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

	// block until told to shutdown
	select {
	case <-g.Options.StopCh:
	case <-g.processExitChan():
		if !g.process.State().Success() {
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

func (g *GenericPilot) processExitChan() <-chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		// don't call wait until the process is actually created
		for {
			// wait until the proces is running before calling Wait
			if g.process != nil && g.process.Running() {
				break
			}
			time.Sleep(2)
		}
		// we must call Wait else process.ProcessState won't be populated with
		// exit code/details of the process.
		g.process.Wait()
	}()
	return out
}
