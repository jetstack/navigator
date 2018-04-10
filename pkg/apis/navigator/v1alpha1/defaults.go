package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
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
	dc := cassDefaultDatacenter
	if np.Datacenter == nil {
		np.Datacenter = &dc
	}

	rack := np.Name
	if np.Rack == nil {
		np.Rack = &rack
	}

	if np.Replicas == nil {
		np.Replicas = ptr.Int32(1)
	}
}

func SetDefaults_ElasticsearchClusterNodePool(np *ElasticsearchClusterNodePool) {
	if np.Replicas == nil {
		np.Replicas = ptr.Int32(1)
	}
}

func SetDefaults_ImageSpec(spec *ImageSpec) {
	if spec.PullPolicy == "" {
		spec.PullPolicy = corev1.PullIfNotPresent
	}
}
