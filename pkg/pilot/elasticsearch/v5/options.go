package v5

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	informers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot"
)

const (
	defaultResyncPeriod = time.Second * 30
	defaultConfigDir    = "/etc/pilot"

	defaultElasticsearchBinary       = "elasticsearch"
	defaultElasticsearchPluginBinary = "elasticsearch-plugin"
	defaultElasticsearchConfigDir    = "/usr/share/elasticsearch/config"
)

// PilotOptions are the options required to run this Pilot. This can be used to
// instantiate new instances of the Pilot.
type PilotOptions struct {
	// RestConfig is the clientset configuration to connection to the apiserver
	// If not specified, autoconfiguration will be performed using the provided
	// master address and kubeconfig path. If these are not specified,
	// in-cluster configuration will be attempted to be loaded.
	RestConfig *rest.Config

	// Master is an optional API server address
	Master string
	// KubeConfig is an optional path to a kubeconfig file
	KubeConfig string
	// ResyncPeriod is how often the controllers should resync their caches
	ResyncPeriod time.Duration
	// ConfigDir is the base directory for additional Pilot configuration
	ConfigDir string

	// ElasticsearchOptions contains the elasticsearch-specific options
	ElasticsearchOptions *ElasticsearchOptions
	// GenericPilotOptions contains options for the genericpilot
	GenericPilotOptions *genericpilot.Options

	StdOut io.Writer
	StdErr io.Writer

	pilot                 *Pilot
	kubeClientset         kubernetes.Interface
	navigatorClientset    clientset.Interface
	sharedInformerFactory informers.SharedInformerFactory
}

type ElasticsearchOptions struct {
	MasterURL string
	// Binary is the name of the elasticsearch binary
	Binary string
	// PluginBinary is the name of the binary used to install plugins
	PluginBinary string
	// ConfigDir is the path to the elasticsearch config directory
	ConfigDir string
	// Roles is a list of ElasticsearchCluster roles this Pilot should use
	Roles []string
	// Plugins is a list of plugins to install for elasticsearch is started
	Plugins []string
}

func NewOptions(out, errOut io.Writer) *PilotOptions {
	o := &PilotOptions{
		ElasticsearchOptions: &ElasticsearchOptions{},
		GenericPilotOptions:  &genericpilot.Options{},
		StdOut:               out,
		StdErr:               errOut,
	}
	return o
}

func (o *PilotOptions) AddFlags(flags *pflag.FlagSet) {
	o.ElasticsearchOptions.AddFlags(flags)
	o.GenericPilotOptions.AddFlags(flags)

	flags.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flags.StringVar(&o.Master, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flags.DurationVar(&o.ResyncPeriod, "resync-period", defaultResyncPeriod, "Re-sync period for control loops operated by the pilot")
	flags.StringVar(&o.ConfigDir, "config-dir", defaultConfigDir, "Base directory for additional Pilot configuration")
}

func (o *PilotOptions) Complete() error {
	var err error
	if o.RestConfig == nil {
		if o.RestConfig, err = clientcmd.BuildConfigFromFlags(o.Master, o.KubeConfig); err != nil {
			return err
		}
	}
	o.kubeClientset, err = kubernetes.NewForConfig(o.RestConfig)
	if err != nil {
		return err
	}
	o.navigatorClientset, err = clientset.NewForConfig(o.RestConfig)
	if err != nil {
		return err
	}
	o.sharedInformerFactory = informers.NewFilteredSharedInformerFactory(o.navigatorClientset, o.ResyncPeriod, o.GenericPilotOptions.PilotNamespace, nil)

	// NewPilot sets some fields on the GenericControllerOptions
	if o.pilot, err = NewPilot(o); err != nil {
		return err
	}

	o.GenericPilotOptions.KubernetesClient = o.kubeClientset
	o.GenericPilotOptions.NavigatorClient = o.navigatorClientset
	o.GenericPilotOptions.SharedInformerFactory = o.sharedInformerFactory
	o.GenericPilotOptions.CmdFunc = o.pilot.CmdFunc
	o.GenericPilotOptions.SyncFunc = o.pilot.syncFunc
	o.GenericPilotOptions.Hooks = o.pilot.Hooks()
	o.GenericPilotOptions.ReadinessProbe = o.pilot.ReadinessCheck()
	o.GenericPilotOptions.LivenessProbe = o.pilot.LivenessCheck()

	if err := o.GenericPilotOptions.Complete(); err != nil {
		return err
	}
	if err := o.ElasticsearchOptions.Complete(); err != nil {
		return err
	}
	return nil
}

func (o *PilotOptions) Validate() error {
	var errs []error
	errs = append(errs, o.ElasticsearchOptions.Validate()...)
	errs = append(errs, o.GenericPilotOptions.Validate()...)
	return utilerrors.NewAggregate(errs)
}

func (o *PilotOptions) Run(stopCh <-chan struct{}) error {
	genericPilot, err := o.GenericPilotOptions.Pilot()
	if err != nil {
		return err
	}

	// set the genericPilot on
	o.pilot.genericPilot = genericPilot

	// create a new stopCh just for the factory so the factory continues to
	// receive updates after the process has been signaled to exit. This allows
	// the Pilot to properly interact with the apiserver whilst it is shutting
	// down, and ensures that the shared informers only stop once the process
	// is ready to exit.
	stopInformers := make(chan struct{})
	defer close(stopInformers)
	// start the shared informer factory
	go o.sharedInformerFactory.Start(stopInformers)

	if err := o.pilot.WaitForCacheSync(stopCh); err != nil {
		return err
	}
	if err := genericPilot.WaitForCacheSync(stopCh); err != nil {
		return err
	}

	return genericPilot.Run()
}

func (e *ElasticsearchOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&e.MasterURL, "elasticsearch-master-url", "", "URL of the Elasticsearch master service")
	flags.StringVar(&e.Binary, "elasticsearch-binary", defaultElasticsearchBinary, "Path to the elasticsearch binary")
	flags.StringVar(&e.PluginBinary, "elasticsearch-plugin-binary", defaultElasticsearchPluginBinary, "Path to the elasticsearch-plugin binary")
	flags.StringVar(&e.ConfigDir, "elasticsearch-config-dir", defaultElasticsearchConfigDir, "Path to the elasticsearch config directory")
	flags.StringSliceVar(&e.Roles, "elasticsearch-roles", nil, "List of roles this node should assume within the Elasticsearch cluster")
	flags.StringSliceVar(&e.Plugins, "elasticsearch-plugins", nil, "List of plugins to install before the elasticsearch process is started")
}

func (e *ElasticsearchOptions) Complete() error {
	return nil
}

func (e *ElasticsearchOptions) Validate() []error {
	var errs []error
	if e.MasterURL == "" {
		errs = append(errs, fmt.Errorf("elasticsearch master URL must be specified"))
	}
	if e.Binary == "" {
		errs = append(errs, fmt.Errorf("elasticsearch binary must be specified"))
	}
	if e.PluginBinary == "" {
		errs = append(errs, fmt.Errorf("elasticsearch plugin binary must be specified"))
	}
	return errs
}
