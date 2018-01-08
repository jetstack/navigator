package genericpilot

import "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"

// IsThisPilot will return true if 'pilot' corresponds to the Pilot resource
// for this pilot.
func (g *GenericPilot) IsThisPilot(pilot *v1alpha1.Pilot) bool {
	return g.isThisPilot(pilot.Name, pilot.Namespace)
}

func (g *GenericPilot) isThisPilot(name, namespace string) bool {
	return name == g.Options.PilotName && namespace == g.Options.PilotNamespace
}
