package configmap

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

const configTemplate = `
node.data: ${NODE_DATA}
node.master: ${NODE_MASTER}
node.ingest: ${NODE_INGEST}
node.name: ${POD_NAME}

network.host: 0.0.0.0

discovery:
  zen:
    ping.unicast.hosts: ${DISCOVERY_URL}
    minimum_master_nodes: %d

xpack.security.enabled: false
`

func esConfigConfigMap(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.ConfigMapName(c, np),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
			Labels:          util.ClusterLabels(c),
		},
		// TODO: move the environment variable names into a general purpose package
		Data: map[string]string{
			"elasticsearch.yml": generateConfig(c, np),
		},
	}
}

func generateConfig(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) string {
	return fmt.Sprintf(configTemplate, c.Spec.MinimumMasters)
}
