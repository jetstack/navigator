package service

import (
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
) *apiv1.Service {
	return UpdateServiceForCluster(cluster, &apiv1.Service{})
}

func UpdateServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
	service *apiv1.Service,
) *apiv1.Service {
	service = service.DeepCopy()
	service.SetName(util.ResourceBaseName(cluster))
	service.SetNamespace(cluster.Namespace)
	service.SetLabels(util.ClusterLabels(cluster))
	service.SetOwnerReferences([]metav1.OwnerReference{
		util.NewControllerRef(cluster),
	})
	service.Spec.Type = apiv1.ServiceTypeClusterIP
	service.Spec.Ports = []apiv1.ServicePort{
		{
			Name:       "transport",
			Port:       int32(9042),
			TargetPort: intstr.FromInt(9042),
		},
	}
	service.Spec.Selector = util.NodePoolLabels(cluster, "")
	return service
}
