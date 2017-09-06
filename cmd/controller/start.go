package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	intclient "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	_ "github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch"
	"github.com/jetstack-experimental/navigator/pkg/kube"
)

type NavigatorControllerOptions struct {
	ControllerOptions *ControllerOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewNavigatorControllerOptions(out, errOut io.Writer) *NavigatorControllerOptions {
	o := &NavigatorControllerOptions{
		ControllerOptions: &ControllerOptions{},

		StdOut: out,
		StdErr: errOut,
	}

	return o
}

// NewCommandStartNavigatorController is a CLI handler for starting navigator
func NewCommandStartNavigatorController(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := NewNavigatorControllerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:   "navigator",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,

		// TODO: Refactor this function from this package
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunNavigatorController(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.ControllerOptions.AddFlags(flags)

	return cmd
}

func (o NavigatorControllerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.ControllerOptions.Validate())
	return utilerrors.NewAggregate(errors)
}

func (o *NavigatorControllerOptions) Complete() error {
	return nil
}

func (o NavigatorControllerOptions) Context() (*controllers.Context, error) {
	// Load the users Kubernetes config
	cfg, err := kube.Config(o.ControllerOptions.APIServerHost)

	if err != nil {
		return nil, fmt.Errorf("error creating rest config: %s", err.Error())
	}

	// Create a Navigator api client
	intcl, err := intclient.NewForConfig(cfg)

	if err != nil {
		return nil, fmt.Errorf("error creating internal group client: %s", err.Error())
	}

	// Create a Kubernetes api client
	cl, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %s", err.Error())
	}

	// Create a context for controllers to use
	ctx := &controllers.Context{
		Client:                cl,
		NavigatorClient:       intcl,
		SharedInformerFactory: kube.NewSharedInformerFactory(),

		Namespace: o.ControllerOptions.Namespace,
	}

	return ctx, nil
}

func (o NavigatorControllerOptions) RunNavigatorController(stopCh <-chan struct{}) error {
	ctx, err := o.Context()
	if err != nil {
		return err
	}
	// Start all known controller loops
	return controllers.Start(ctx, controllers.Known(), stopCh)
}
