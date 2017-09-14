package main

import (
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jetstack-experimental/navigator/cmd/controller/app"
	"github.com/jetstack-experimental/navigator/cmd/controller/app/options"
	_ "github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch"
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

// NewCommandStartCertManagerController is a CLI handler for starting cert-manager
func NewCommandStartNavigatorController(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := NewNavigatorControllerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:   "cert-manager-controller",
		Short: "Automated TLS controller for Kubernetes",
		Long: `
cert-manager is a Kubernetes addon to automate the management and issuance of
TLS certificates from various issuing sources.

It will ensure certificates are valid and up to date periodically, and attempt
to renew certificates at an appropriate time before expiry.`,

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
