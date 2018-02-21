package v3

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// hack to test the seedprovider, this should use whatever pattern is decided upon here:
	//   https://github.com/jetstack/navigator/issues/251
	cfgPath := "/etc/cassandra/cassandra.yaml"
	read, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	newContents := strings.Replace(string(read),
		"org.apache.cassandra.locator.SimpleSeedProvider",
		"io.k8s.cassandra.KubernetesSeedProvider", -1)

	newContents = strings.Replace(newContents,
		"SimpleSnitch",
		"GossipingPropertyFileSnitch", -1)

	err = ioutil.WriteFile(cfgPath, []byte(newContents), 0)
	if err != nil {
		return nil, err
	}

	snitchSettings, err := p.getSnitchSettings()
	if err != nil {
		return nil, err
	}

	glog.V(4).Info("generated snitch: %v", snitchSettings)

	err = ioutil.WriteFile("/cassandra-rackdc.properties", []byte(snitchSettings), 0)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pilot) getSnitchSettings() (string, error) {
	nodeClientset := p.Options.kubeClientset.CoreV1().Nodes()

	nodeName := os.Getenv("NODE_NAME")
	node, err := nodeClientset.Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	datacenter := "navigator-datacenter-1"
	rack := "navigator-rack-1"

	if p.Options.FailureZoneDatacenterNodeLabel != "" {
		if val, ok := node.Labels[p.Options.FailureZoneDatacenterNodeLabel]; ok {
			datacenter = val
		}
	}

	if p.Options.FailureZoneRackNodeLabel != "" {
		if val, ok := node.Labels[p.Options.FailureZoneRackNodeLabel]; ok {
			rack = val
		}
	}

	return fmt.Sprintf("dc=%s\nrack=%s", datacenter, rack), nil
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
