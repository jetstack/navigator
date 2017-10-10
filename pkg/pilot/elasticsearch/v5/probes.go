package v5

import (
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/probe"
)

func (p *Pilot) ReadinessCheck() probe.Check {
	return probe.CombineChecks(
		func() error {
			return nil
		},
	)
}

func (p *Pilot) LivenessCheck() probe.Check {
	return probe.CombineChecks(
		func() error {
			return nil
		},
	)
}
