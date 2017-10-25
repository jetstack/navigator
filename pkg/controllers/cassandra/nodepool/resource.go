package nodepool

import (
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"k8s.io/api/apps/v1beta2"
)

func StatefulSetForCluster(
	cluster *v1alpha1.CassandraCluster,
	nodepool *v1alpha1.CassandraClusterNodePool,
) *v1beta2.StatefulSet {
	set := &v1beta2.StatefulSet{}
	set.SetName(nodepool.Name)
	set.SetNamespace(cluster.Namespace)
	return set
}
