package v5

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

func (p *Pilot) actionDecommission(pilot *v1alpha1.Pilot) error {
	// find Decommissioned condition
	for _, cond := range pilot.Status.Conditions {
		if cond.Type != v1alpha1.PilotConditionDecommissioned {
			continue
		}

	}
	return nil
}

func (p *Pilot) periodicDecommission(pilot *v1alpha1.Pilot) error {
	if pilot.Spec.Phase != v1alpha1.PilotPhaseDecommissioned {
		return nil
	}
	return nil
}
