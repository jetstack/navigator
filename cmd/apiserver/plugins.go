package main

import (
	"k8s.io/apiserver/pkg/admission"

	// Admission controllers
	"github.com/jetstack/navigator/plugin/pkg/admission/namespace/lifecycle"
)

// registerAllAdmissionPlugins registers all admission plugins
func registerAllAdmissionPlugins(plugins *admission.Plugins) {
	lifecycle.Register(plugins)
}
