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

	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	informers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot"
)

const (
	defaultResyncPeriod = time.Second * 30

	defaultBinary       = "elasticsearch"
	defaultPluginBinary = "elasticsearch-plugin"
	defaultConfigDir    = "/usr/share/elasticsearch/config"
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
	o.sharedInformerFactory = informers.NewSharedInformerFactory(o.navigatorClientset, o.ResyncPeriod)

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

	// start the shared informer factories
	go o.sharedInformerFactory.Start(stopCh)

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
	flags.StringVar(&e.Binary, "elasticsearch-binary", defaultBinary, "Path to the elasticsearch binary")
	flags.StringVar(&e.PluginBinary, "elasticsearch-plugin-binary", defaultPluginBinary, "Path to the elasticsearch-plugin binary")
	flags.StringVar(&e.ConfigDir, "elasticsearch-config-dir", defaultConfigDir, "Path to the elasticsearch config directory")
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
