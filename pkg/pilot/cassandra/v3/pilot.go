package v3

import (
	"fmt"
	"os/exec"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
)

type Pilot struct {
	Options *PilotOptions

	// navigator clientset
	navigatorClient clientset.Interface

	pilotLister         listersv1alpha1.PilotLister
	pilotInformerSynced cache.InformerSynced
	// a reference to the GenericPilot for this Pilot
	genericPilot *genericpilot.GenericPilot
}

func NewPilot(opts *PilotOptions) (*Pilot, error) {
	pilotInformer := opts.sharedInformerFactory.Navigator().V1alpha1().Pilots()

	p := &Pilot{
		Options:             opts,
		navigatorClient:     opts.navigatorClientset,
		pilotLister:         pilotInformer.Lister(),
		pilotInformerSynced: pilotInformer.Informer().HasSynced,
	}

	return p, nil
}

func (p *Pilot) WaitForCacheSync(stopCh <-chan struct{}) error {
	if !cache.WaitForCacheSync(stopCh, p.pilotInformerSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	return nil
}

func (p *Pilot) Hooks() *hook.Hooks {
	return &hook.Hooks{}
}

func (p *Pilot) CmdFunc(pilot *v1alpha1.Pilot) (*exec.Cmd, error) {
	// The /run.sh script is unique to gcr.io/google-samples/cassandra:v12.
	// TODO: Add support for other Cassandra images with different entry points.
	cmd := exec.Command("/run.sh")
	return cmd, nil
}

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	if p.genericPilot.IsThisPilot(pilot) {
	}
	return nil
}

func (p *Pilot) ReadinessCheck() error {
	glog.V(2).Infof("readiness status: %q", "ok")
	return nil
}

func (p *Pilot) LivenessCheck() error {
	glog.V(2).Infof("liveness status: %q", "ok")
	return nil
}
