package cql

import (
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	serviceutil "github.com/jetstack/navigator/pkg/controllers/cassandra/service/util"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
) *apiv1.Service {
	return updateServiceForCluster(cluster, &apiv1.Service{})
}

func updateServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
	service *apiv1.Service,
) *apiv1.Service {
	service = service.DeepCopy()
	service = serviceutil.SetStandardServiceAttributes(cluster, service)
	service.SetName(util.CqlServiceName(cluster))
	service.Spec.Type = apiv1.ServiceTypeClusterIP
	service.Spec.Ports = []apiv1.ServicePort{
		{
			Name:       "cql",
			Port:       cluster.Spec.CqlPort,
			TargetPort: intstr.FromInt(util.DefaultCqlPort),
		},
	}
	return service
}
