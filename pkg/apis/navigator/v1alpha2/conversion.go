package v1alpha2

import (
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
)

func addConversionFuncs(scheme *runtime.Scheme) error {
	// Add non-generated conversion functions
	return scheme.AddConversionFuncs(
		Convert_v1alpha2_ElasticsearchClusterNodePool_To_navigator_ElasticsearchClusterNodePool,
		Convert_navigator_ElasticsearchClusterNodePool_To_v1alpha2_ElasticsearchClusterNodePool,
		Convert_navigator_ElasticsearchClusterPersistenceConfig_To_v1alpha2_ElasticsearchClusterPersistenceConfig,
	)
}

func Convert_navigator_ElasticsearchClusterPersistenceConfig_To_v1alpha2_ElasticsearchClusterPersistenceConfig(in *navigator.ElasticsearchClusterPersistenceConfig, out *ElasticsearchClusterPersistenceConfig, s conversion.Scope) error {
	return autoConvert_navigator_ElasticsearchClusterPersistenceConfig_To_v1alpha2_ElasticsearchClusterPersistenceConfig(in, out, s)
}

func Convert_navigator_ElasticsearchClusterNodePool_To_v1alpha2_ElasticsearchClusterNodePool(in *navigator.ElasticsearchClusterNodePool, out *ElasticsearchClusterNodePool, s conversion.Scope) error {
	if in.Persistence.Enabled {
		out.Persistence = &ElasticsearchClusterPersistenceConfig{}
		if err := Convert_navigator_ElasticsearchClusterPersistenceConfig_To_v1alpha2_ElasticsearchClusterPersistenceConfig(&in.Persistence, out.Persistence, s); err != nil {
			return err
		}
	}
	return autoConvert_navigator_ElasticsearchClusterNodePool_To_v1alpha2_ElasticsearchClusterNodePool(in, out, s)
}

func Convert_v1alpha2_ElasticsearchClusterNodePool_To_navigator_ElasticsearchClusterNodePool(in *ElasticsearchClusterNodePool, out *navigator.ElasticsearchClusterNodePool, s conversion.Scope) error {
	if in.Persistence != nil {
		out.Persistence.Enabled = true
		if err := Convert_v1alpha2_ElasticsearchClusterPersistenceConfig_To_navigator_ElasticsearchClusterPersistenceConfig(in.Persistence, &out.Persistence, s); err != nil {
			return err
		}
	}
	return autoConvert_v1alpha2_ElasticsearchClusterNodePool_To_navigator_ElasticsearchClusterNodePool(in, out, s)
}
