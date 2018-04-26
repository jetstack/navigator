package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/jetstack/navigator/pkg/pilot/cassandra/v3"
)

// NewCommandStartPilot provides a CLI handler for the pilot
func NewCommandStartPilot(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := v3.NewOptions(out, errOut)

	cmd := &cobra.Command{
		Use:   "navigator-pilot-cassandra",
		Short: "Launch a Navigator Cassandra Pilot",
		Long: `
Launch a Navigator Cassandra Pilot.

Navigator is a Kubernetes extension for managing common stateful services on Kubernetes.
Documentation is available at https://navigator-dbaas.readthedocs.io.
`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(stopCh); err != nil {
				return err
			}
			return nil
		},
	}
	flags := cmd.Flags()
	o.AddFlags(flags)
	o.GenericPilotOptions.StopCh = stopCh

	return cmd
}
