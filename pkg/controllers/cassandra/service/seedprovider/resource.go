package seedprovider

import (
	apiv1 "k8s.io/api/core/v1"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	serviceutil "github.com/jetstack/navigator/pkg/controllers/cassandra/service/util"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

const (
	TolerateUnreadyEndpointsAnnotationKey = "service.alpha.kubernetes.io/tolerate-unready-endpoints"
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
	service.SetName(util.SeedProviderServiceName(cluster))
	service.Spec.Type = apiv1.ServiceTypeClusterIP
	service.Spec.ClusterIP = "None"
	// Headless service should not require a port.
	// But without it, DNS records are not registered.
	// See https://github.com/kubernetes/kubernetes/issues/55158
	service.Spec.Ports = []apiv1.ServicePort{
		{
			Port: 65535,
		},
	}
	// This ensures that DNS names are published regardless of whether the
	// Cassandra pod ReadinessProbes (listening on their CQL port).
	// It won't handle CQL connections until it has successfully connected and
	// negotiated with a seed host
	service.Spec.PublishNotReadyAddresses = true
	// XXX: This annotation is only necessary for Kubernetes <=1.7, which do not
	// pay attention to the field above.
	// Remove it when we no longer support those versions.
	annotations := service.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[TolerateUnreadyEndpointsAnnotationKey] = "true"
	service.SetAnnotations(annotations)
	return service
}
