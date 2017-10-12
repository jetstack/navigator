package configmap

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

func esConfigConfigMap(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.ConfigMapName(c, np),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
			Labels:          util.ClusterLabels(c),
		},
		Data: np.Config,
	}
}
