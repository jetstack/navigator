package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/pkg/util/ptr"
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
	if np.Seeds == nil {
		np.Seeds = ptr.Int32(1)
	}
}
