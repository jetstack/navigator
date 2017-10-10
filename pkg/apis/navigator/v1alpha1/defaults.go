package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&Pilot{}, pilotDefaultFunc)
	return RegisterDefaults(scheme)
}

func pilotDefaultFunc(obj interface{}) {
	// this is a safe cast as AddTypeDefaultingFunc states that the function
	// will not be called with an object of type other than the srcType
	pilot := obj.(*Pilot)
	if pilot.Spec.Phase == "" {
		pilot.Spec.Phase = PilotPhaseInitializing
	}
}
