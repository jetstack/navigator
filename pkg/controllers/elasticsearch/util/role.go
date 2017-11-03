package util

import "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"

func RBACRoleName(c *v1alpha1.ElasticsearchCluster) string {
	return ResourceBaseName(c)
}
