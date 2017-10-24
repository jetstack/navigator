package service

import (
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/util"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterServiceControl struct {
	kubeClient    kubernetes.Interface
	serviceLister corelisters.ServiceLister
}

var _ Interface = &defaultCassandraClusterServiceControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	serviceLister corelisters.ServiceLister,
) Interface {
	return &defaultCassandraClusterServiceControl{
		kubeClient:    kubeClient,
		serviceLister: serviceLister,
	}
}

func (e *defaultCassandraClusterServiceControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	svc := ServiceForCluster(cluster)
	_, err := e.kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func ServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
) *apiv1.Service {
	svc := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.ResourceBaseName(cluster),
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster),
			Annotations:     make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
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
		},
	}
	return &svc
}
