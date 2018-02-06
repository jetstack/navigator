package v3

import (
	"fmt"
	"os"
	"os/exec"

	"k8s.io/client-go/tools/cache"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/cassandra/nodetool"
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
	nodeTool     nodetool.Interface
}

func NewPilot(opts *PilotOptions) (*Pilot, error) {
	pilotInformer := opts.sharedInformerFactory.Navigator().V1alpha1().Pilots()

	p := &Pilot{
		Options:             opts,
		navigatorClient:     opts.navigatorClientset,
		pilotLister:         pilotInformer.Lister(),
		pilotInformerSynced: pilotInformer.Informer().HasSynced,
		nodeTool:            opts.nodeTool,
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
	cmd := exec.Command("/docker-entrypoint.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	if p.genericPilot.IsThisPilot(pilot) {
	}
	return nil
}

func localNodeUpAndNormal(nodeTool nodetool.Interface) error {
	nodes, err := nodeTool.Status()
	if err != nil {
		return err
	}
	localNode := nodes.LocalNode()
	if localNode == nil {
		return fmt.Errorf("Local node not found: %v", nodes)
	}
	if localNode.Status != nodetool.NodeStatusUp {
		return fmt.Errorf("Unexpected local node status: %v", localNode.Status)
	}
	if localNode.State != nodetool.NodeStateNormal {
		return fmt.Errorf("Unexpected local node state: %v", localNode.State)
	}
	return nil
}

// If a node is recovering, or still starting up, the readiness probe should fail
// but liveness should pass.
// If a liveness probe fails, Kubernetes will begin restarting that pod,
// which can quite easily cause the pod to never start as a result of constant CrashLoopBackOff
// So the liveness probe here should do enough to demonstrate that the Pilot is alive
// and that the Cassandra is responding to JMX / Jolokia HTTP requests.
// TODO: The Readiness probe should also attempt to make a CQL connection.

func (p *Pilot) ReadinessCheck() error {
	return localNodeUpAndNormal(p.nodeTool)
}

func (p *Pilot) LivenessCheck() error {
	_, err := p.nodeTool.Status()
	return err
}
