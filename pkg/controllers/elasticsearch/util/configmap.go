package util

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func ConfigMapName(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) string {
	return NodePoolResourceName(c, np)
}
