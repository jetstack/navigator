package controller

import (
	"fmt"

	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

// IsThisPilot will return true if 'pilot' corresponds to the Pilot resource
// for this pilot.
func (g *Controller) IsThisPilot(pilot *v1alpha1.Pilot) bool {
	return g.isThisPilot(pilot.Name, pilot.Namespace)
}

func (g *Controller) isThisPilot(name, namespace string) bool {
	return name == g.pilotName && namespace == g.pilotNamespace
}

func (g *Controller) IsPeer(pilot *v1alpha1.Pilot) (bool, error) {
	// get a reference to 'this' pilot
	thisPilot, err := g.ThisPilot()
	if err != nil {
		return false, err
	}

	clusterOwnerRef := metav1.GetControllerOf(thisPilot)
	if clusterOwnerRef == nil {
		return false, fmt.Errorf("cannot determine owner of this Pilot resource (%q) as it is nil. this is an invalid state", g.pilotName)
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

// ThisPilot will return a reference to 'this' Pilot resource. The returned
// resource may or may not be up to date, and it may or may not still exist in
// the target API server.
func (g *Controller) ThisPilot() (*v1alpha1.Pilot, error) {
	// get a reference to 'this' pilot
	thisPilot, err := g.pilotLister.Pilots(g.pilotNamespace).Get(g.pilotName)
	if apierrors.IsNotFound(err) {
		if g.cachedThisPilot != nil {
			return g.cachedThisPilot, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return thisPilot, nil
}
