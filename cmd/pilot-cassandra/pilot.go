package main

import (
	"io"

	"github.com/jetstack/navigator/pkg/pilot/cassandra/v3"
	"github.com/spf13/cobra"
)

// NewCommandStartPilot provides a CLI handler for the pilot
func NewCommandStartPilot(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := v3.NewOptions(out, errOut)

	cmd := &cobra.Command{
		Short: "Launch a Cassandra Pilot",
		Long:  "Launch a Cassandra Pilot",
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
