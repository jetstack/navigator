package genericpilot

import (
	"fmt"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

// IsThisPilot will return true if 'pilot' corresponds to the Pilot resource
// for this pilot.
func (g *GenericPilot) IsThisPilot(pilot *v1alpha1.Pilot) bool {
	return pilot.Name == g.Options.PilotName && pilot.Namespace == g.Options.PilotNamespace
}

func (g *GenericPilot) IsPeer(pilot *v1alpha1.Pilot) (bool, error) {
	// get a reference to 'this' pilot
	thisPilot, err := g.pilotLister.Pilots(g.Options.PilotNamespace).Get(g.Options.PilotName)
	if err != nil {
		return false, err
	}

	clusterOwnerRef := metav1.GetControllerOf(thisPilot)
	if clusterOwnerRef == nil {
		return false, fmt.Errorf("cannot determine owner of this Pilot resource (%q) as it is nil. this is an invalid state", g.Options.PilotName)
	}

	pilotOwnerRef := metav1.GetControllerOf(pilot)
	if pilotOwnerRef == nil {
		glog.V(4).Infof("cannot determine owner of the provided Pilot resource (%q) as it is nil. skipping processing Pilot", pilot.Name)
		return false, nil
	}

	return clusterOwnerRef.Name == pilotOwnerRef.Name &&
		clusterOwnerRef.UID == pilotOwnerRef.UID &&
		clusterOwnerRef.Kind == pilotOwnerRef.Kind &&
		clusterOwnerRef.APIVersion == pilotOwnerRef.APIVersion, nil
}

func (g *GenericPilot) IsRunning() bool {
	return g.process != nil && g.process.State() != nil && g.process.State().Exited() == false
}
