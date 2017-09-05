package service

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

func discoveryService(c *v1alpha1.ElasticsearchCluster) *apiv1.Service {
	svc := buildService(c, util.DiscoveryServiceName(c), false, util.NodePoolLabels(c, ""))
	svc.Annotations["service.alpha.kubernetes.io/tolerate-unready-endpoints"] = "true"
	return svc
}

func clientService(c *v1alpha1.ElasticsearchCluster) *apiv1.Service {
	return buildService(c, util.ClientServiceName(c), false, util.NodePoolLabels(c, "", "client"))
}

func buildService(c *v1alpha1.ElasticsearchCluster, name string, http bool, selector map[string]string) *apiv1.Service {
	svc := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       c.Namespace,
			Labels:          util.ClusterLabels(c),
			Annotations:     make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Name:       "transport",
					Port:       int32(9300),
					TargetPort: intstr.FromInt(9300),
				},
			},
			Selector: selector,
		},
	}

	if http {
		svc.Spec.Ports = append(svc.Spec.Ports, apiv1.ServicePort{
			Name:       "http",
			Port:       int32(9200),
			TargetPort: intstr.FromInt(9200),
		})
	}

	return &svc
}
