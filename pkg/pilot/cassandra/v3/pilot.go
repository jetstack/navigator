package v3

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/cassandra/nodetool"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/config"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
)

const (
	CassSnitch       = "GossipingPropertyFileSnitch"
	CassSeedProvider = "io.jetstack.cassandra.KubernetesSeedProvider"

	CassandraYaml             = "cassandra.yaml"
	CassandraRackDcProperties = "cassandra-rackdc.properties"
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
	return &hook.Hooks{
		PreStart: []hook.Interface{
			hook.New("WriteConfig", p.WriteConfig),
		},
	}
}

func (p *Pilot) CmdFunc(pilot *v1alpha1.Pilot) (*exec.Cmd, error) {
	cmd := exec.Command(p.Options.CassandraPath, "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	if pilot.Status.Cassandra == nil {
		pilot.Status.Cassandra = &v1alpha1.CassandraPilotStatus{}
	}

	version, err := p.nodeTool.Version()
	if err != nil {
		pilot.Status.Cassandra.Version = nil
		glog.Errorf("error while getting Cassandra version: %s", err)
	}
	pilot.Status.Cassandra.Version = version
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

func (p *Pilot) writeCassandraYaml(pilot *v1alpha1.Pilot) error {
	cfg, err := config.NewFromYaml(
		path.Join(p.Options.CassandraConfigPath, "cassandra.yaml"),
	)
	if err != nil {
		return err
	}
	cfg.Set("cluster_name", p.Options.CassandraClusterName)
	// We unset any pre-configured IP / host addresses so that cassandra is
	// forced to lookup of its FQDN and corresponding IP address.
	// This allows cassandra to handle changes of IP address if Pods are restarted.
	cfg.Unset("listen_address")
	cfg.Unset("listen_interface")
	cfg.Unset("broadcast_address")
	cfg.Unset("rpc_address")
	// Force the use of GossipingPropertyFileSnitch so that cassandra will query the
	// rackdc properties file for rack and DC values.
	cfg.Set("endpoint_snitch", CassSnitch)
	// Force the use of Kubernetes seed provider
	// And remove preconfigured seeds so that cassandra can only get seed
	// information from the Kubernetes service.
	cfg.Set(
		"seed_provider",
		[]map[string]interface{}{{
			"class_name": CassSeedProvider,
			"parameters": []interface{}{
				map[string]interface{}{
					"seeds": "",
				},
			},
		}},
	)
	return cfg.WriteConfig()
}

func (p *Pilot) writeCassandraRackDcProperties(pilot *v1alpha1.Pilot) error {
	cfg, err := config.NewFromProperties(
		path.Join(p.Options.CassandraConfigPath, "cassandra-rackdc.properties"),
	)
	if err != nil {
		return err
	}
	cfg.Set("rack", p.Options.CassandraRack)
	cfg.Set("dc", p.Options.CassandraDC)
	return cfg.WriteConfig()
}

// WriteConfig creates Navigator compatible Cassandra configuration files.
// It reads default cassandra.yaml and cassandra-rackdc.properties files
// and overrides configuration values that are vital for Navigator to manage
// the cassandra cluster.
func (p *Pilot) WriteConfig(pilot *v1alpha1.Pilot) (err error) {
	err = p.writeCassandraYaml(pilot)
	if err != nil {
		return errors.Wrap(err, "unable to write cassandra.yaml")
	}

	err = p.writeCassandraRackDcProperties(pilot)
	if err != nil {
		return errors.Wrap(err, "unable to write cassandra-rackdc.properties")
	}

	return nil
}
