package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/jetstack/navigator/pkg/pilot/elasticsearch/v5"
)

// NewCommandStartPilot provides a CLI handler for the pilot
func NewCommandStartPilot(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := v5.NewOptions(out, errOut)

	cmd := &cobra.Command{
		Short: "Launch an Elasticsearch Pilot",
		Long:  "Launch an Elasticsearch Pilot",
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
