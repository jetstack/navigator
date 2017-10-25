package nodepool

import (
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/util"
	"k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/apps"
)

func StatefulSetForCluster(
	cluster *v1alpha1.CassandraCluster,
	np *v1alpha1.CassandraClusterNodePool,
) *v1beta2.StatefulSet {

	statefulSetName := util.NodePoolResourceName(cluster, np)

	set := &v1beta2.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            statefulSetName,
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster),
			Annotations:     make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		Spec: v1beta2.StatefulSetSpec{
			Replicas:    util.Int32Ptr(int32(np.Replicas)),
			ServiceName: statefulSetName,
			Selector: &metav1.LabelSelector{
				MatchLabels: util.NodePoolLabels(c, np.Name),
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: apps.ParallelPodManagement,
			Template:            *elasticsearchPodTemplate,
		},
	}
	return set
}
