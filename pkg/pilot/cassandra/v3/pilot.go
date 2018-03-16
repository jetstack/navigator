package v3

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"

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

	decommissionInProgress bool
}

func NewPilot(opts *PilotOptions) (*Pilot, error) {
	pilotInformer := opts.sharedInformerFactory.Navigator().V1alpha1().Pilots()

	p := &Pilot{
		Options:                opts,
		navigatorClient:        opts.navigatorClientset,
		pilotLister:            pilotInformer.Lister(),
		pilotInformerSynced:    pilotInformer.Informer().HasSynced,
		nodeTool:               opts.nodeTool,
		decommissionInProgress: false,
	}

	// hack to test the seedprovider, this should use whatever pattern is decided upon here:
	//   https://github.com/jetstack/navigator/issues/251
	cfgPath := "/etc/cassandra/cassandra.yaml"
	read, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	newContents := strings.Replace(string(read),
		"org.apache.cassandra.locator.SimpleSeedProvider",
		"io.jetstack.cassandra.KubernetesSeedProvider", -1)

	err = ioutil.WriteFile(cfgPath, []byte(newContents), 0)
	if err != nil {
		return nil, err
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
	if pilot.Status.Cassandra == nil {
		pilot.Status.Cassandra = &v1alpha1.CassandraPilotStatus{}
	}

	if pilot.Spec.Cassandra != nil {
		if pilot.Spec.Cassandra.Decommissioned {
			p.decommissionInProgress = true
			err := p.decommission()
			if err != nil {
				glog.Errorf("error while decommissioning cassandra node: %s", err)
			} else {
				pilot.Status.Cassandra.Decommissioned = true
			}
		}
	}

	version, err := p.nodeTool.Version()
	if err != nil {
		pilot.Status.Cassandra.Version = nil
		glog.Errorf("error while getting Cassandra version: %s", err)
	}
	pilot.Status.Cassandra.Version = version
	return nil
}

func run(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	os.Unsetenv("JVM_OPTS")
	return cmd.Run()
}

func (p *Pilot) decommission() error {
	nodes, err := p.nodeTool.Status()
	if err != nil {
		return err
	}
	localNode := nodes.LocalNode()

	// if node is operational and working normally, decommission node
	if localNode.State == nodetool.NodeStateNormal {
		glog.Info("about to decomission node")
		return run("nodetool", "decommission")
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
	if localNode.State != nodetool.NodeStateNormal && localNode.State != nodetool.NodeStateLeaving {
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
	if p.decommissionInProgress || true {
		glog.Info("decommission in progress, reporting success for liveness")
		return nil
	}
	_, err := p.nodeTool.Status()
	return err
}
