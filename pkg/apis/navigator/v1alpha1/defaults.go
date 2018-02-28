package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	cassDefaultDatacenter = "navigator-default-datacenter"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_CassandraClusterNodePool(np *CassandraClusterNodePool) {
	if np.Datacenter == "" {
		np.Datacenter = cassDefaultDatacenter
	}

	if np.Rack == "" {
		np.Rack = np.Name
	}

	// default to 1 seed if not specified
	if np.Seeds == 0 {
		np.Seeds = 1
	}
}
