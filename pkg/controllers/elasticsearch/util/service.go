package util

import (
	"fmt"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

func DiscoveryServiceName(c *v1alpha1.ElasticsearchCluster) string {
	return fmt.Sprintf("%s-discovery", ResourceBaseName(c))
}

func ClientServiceName(c *v1alpha1.ElasticsearchCluster) string {
	return fmt.Sprintf("%s", ResourceBaseName(c))
}
