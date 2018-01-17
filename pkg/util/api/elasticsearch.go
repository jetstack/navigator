package api

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func CountElasticsearchMasters(pools []v1alpha1.ElasticsearchClusterNodePool) int64 {
	masters := int64(0)
	for _, pool := range pools {
		if ContainsElasticsearchRole(pool.Roles, v1alpha1.ElasticsearchRoleMaster) {
			masters += pool.Replicas
		}
	}
	return masters
}

func ContainsElasticsearchRole(set []v1alpha1.ElasticsearchClusterRole, role v1alpha1.ElasticsearchClusterRole) bool {
	for _, s := range set {
		if s == role {
			return true
		}
	}
	return false
}
