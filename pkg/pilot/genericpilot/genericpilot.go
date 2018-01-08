package genericpilot

import (
	"sync"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
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
}

func (g *GenericPilot) Run() error {
	glog.Infof("Starting generic pilot controller")

	// setup healthz handlers
	g.serveHealthz()
	ctrlStopCh := make(chan struct{})
	defer close(ctrlStopCh)
	go g.controller.Run(ctrlStopCh)

	// block until told to shutdown
	<-g.Options.StopCh
	glog.V(4).Infof("Shutdown signal received")
	thisPilot, err := g.controller.ThisPilot()
	if err != nil {
		return err
	}
	return g.stop(thisPilot)
}

func (g *GenericPilot) IsRunning() bool {
	if g.process == nil || !g.process.Running() {
		return false
	}
	return true
}
