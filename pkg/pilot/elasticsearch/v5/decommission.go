package v5

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

// actionDecommission will set the shard allocation exclude parameter to
// include the provided Pilot.
func (p *Pilot) actionDecommission(pilot *v1alpha1.Pilot) error {
	// find Decommissioned condition
	for _, cond := range pilot.Status.Conditions {
		if cond.Type != v1alpha1.PilotConditionDecommissioned {
			continue
		}

	}
	return nil
}

// periodicDecommission will update the status field of the given pilot to
// reflect the current decommissioning status. It does this by querying the
// Elasticsearch API to check the exclude parameter, as well as checking how
// many shards are still allocated to the provided node.
func (p *Pilot) periodicDecommission(pilot *v1alpha1.Pilot) error {
	if pilot.Spec.Phase != v1alpha1.PilotPhaseDecommissioned {
		return nil
	}
	return nil
}
