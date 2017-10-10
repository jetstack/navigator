package genericpilot

import "github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/probe"

func (g *GenericPilot) serveHealthz() {
	// Start readiness checker
	go (&probe.Listener{
		Port:  12001,
		Check: g.Options.ReadinessProbe,
	}).Listen()

	// Start liveness checker
	go (&probe.Listener{
		Port:  12000,
		Check: g.Options.LivenessProbe,
	}).Listen()
}
