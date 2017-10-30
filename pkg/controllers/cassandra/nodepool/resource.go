package nodepool

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	apps "k8s.io/api/apps/v1beta2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func StatefulSetForCluster(
	cluster *v1alpha1.CassandraCluster,
	np *v1alpha1.CassandraClusterNodePool,
) *apps.StatefulSet {

	statefulSetName := util.NodePoolResourceName(cluster, np)
	nodePoolLabels := util.NodePoolLabels(cluster, np.Name)
	set := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            statefulSetName,
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster),
			Annotations:     make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    util.Int32Ptr(int32(np.Replicas)),
			ServiceName: statefulSetName,
			Selector: &metav1.LabelSelector{
				MatchLabels: nodePoolLabels,
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: apps.ParallelPodManagement,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: nodePoolLabels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "cassandra",
							Image:           "gcr.io/google-samples/cassandra:v12",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "intra-node",
									ContainerPort: int32(7000),
								},
								{
									Name:          "intra-node-tls",
									ContainerPort: int32(7001),
								},
								{
									Name:          "jmx",
									ContainerPort: int32(7199),
								},
								{
									Name:          "cql",
									ContainerPort: int32(9042),
								},
							},
						},
					},
				},
			},
		},
	}
	return set
}
