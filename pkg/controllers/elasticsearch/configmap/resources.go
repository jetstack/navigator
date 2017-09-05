package configmap

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

func esConfigConfigMap(c *v1alpha1.ElasticsearchCluster) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.ConfigMapName(c),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
			Labels:          util.ClusterLabels(c),
		},
		Data: map[string]string{
			"elasticsearch.yml": `
node.data: ${NODE_DATA}
node.master: ${NODE_MASTER}
node.ingest: ${NODE_INGEST}
node.name: ${HOSTNAME}

network.host: 0.0.0.0

discovery:
  zen:
    ping.unicast.hosts: ${DISCOVERY_SERVICE}
    minimum_master_nodes: 1

xpack.security.enabled: false
`,
			"log4j2.properties": `
status = error

appender.console.type = Console
appender.console.name = console
appender.console.layout.type = PatternLayout
appender.console.layout.pattern = [%d{ISO8601}][%-5p][%-25c{1.}] %marker%m%n

rootLogger.level = debug
rootLogger.appenderRef.console.ref = console
`,
		},
	}
}
