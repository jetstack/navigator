package main

import (
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jetstack/navigator/cmd/controller/app"
	"github.com/jetstack/navigator/cmd/controller/app/options"
	_ "github.com/jetstack/navigator/pkg/controllers/cassandra"
	_ "github.com/jetstack/navigator/pkg/controllers/elasticsearch"
)

type NavigatorControllerOptions struct {
	ControllerOptions *options.ControllerOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewNavigatorControllerOptions(out, errOut io.Writer) *NavigatorControllerOptions {
	o := &NavigatorControllerOptions{
		ControllerOptions: options.NewControllerOptions(),

		StdOut: out,
		StdErr: errOut,
	}

	return o
}

// NewCommandStartNavigatorController provides a CLI handler for the 'navigator-controller' command
func NewCommandStartNavigatorController(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := NewNavigatorControllerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:   "navigator-controller",
		Short: "Launch a Navigator controller",
		Long: `
Launch a Navigator controller.

Navigator is a Kubernetes extension for managing common stateful services on Kubernetes.
Documentation is available at https://navigator-dbaas.readthedocs.io.
`,

		// TODO: Refactor this function from this package
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(args); err != nil {
				glog.Fatalf("error validating options: %s", err.Error())
			}
			o.RunNavigatorController(stopCh)
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

func (o NavigatorControllerOptions) RunNavigatorController(stopCh <-chan struct{}) {
	app.Run(o.ControllerOptions, stopCh)
}
