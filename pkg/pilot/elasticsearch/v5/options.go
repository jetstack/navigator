package v5

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot"
)

const (
	defaultBinary       = "elasticsearch"
	defaultPluginBinary = "elasticsearch-plugin"
)

type Options struct {
	MasterURL string
	// Binary is the name of the elasticsearch binary
	Binary string
	// PluginBinary is the name of the binary used to install plugins
	PluginBinary string
	// GenericPilotOptions contains options for the genericpilot
	GenericPilotOptions *genericpilot.Options

	StdOut io.Writer
	StdErr io.Writer
}

func NewOptions(out, errOut io.Writer) *Options {
	o := &Options{
		GenericPilotOptions: &genericpilot.Options{},
		StdOut:              out,
		StdErr:              errOut,
	}
	return o
}

func (o *Options) AddFlags(flags *pflag.FlagSet) {
	o.GenericPilotOptions.AddFlags(flags)
	flags.StringVar(&o.MasterURL, "elasticsearch-master-url", "", "URL of the Elasticsearch master service")
	flags.StringVar(&o.Binary, "elasticsearch-binary", defaultBinary, "Path to the elasticsearch binary")
	flags.StringVar(&o.PluginBinary, "elasticsearch-plugin-binary", defaultPluginBinary, "Path to the elasticsearch-plugin binary")
}

func (o *Options) Complete() error {
	ConfigureGenericPilot(o)
	return o.GenericPilotOptions.Complete()
}

func (o *Options) Validate() error {
	var errs []error
	if o.MasterURL == "" {
		errs = append(errs, fmt.Errorf("elasticsearch master URL must be specified"))
	}
	if o.Binary == "" {
		errs = append(errs, fmt.Errorf("elasticsearch binary must be specified"))
	}
	if o.PluginBinary == "" {
		errs = append(errs, fmt.Errorf("elasticsearch plugin binary must be specified"))
	}
	errs = append(errs, o.GenericPilotOptions.Validate()...)
	return utilerrors.NewAggregate(errs)
}

func (o *Options) Run(stopCh <-chan struct{}) error {
	genericPilot, err := o.GenericPilotOptions.Pilot()
	if err != nil {
		return err
	}

	return genericPilot.Run()
}
