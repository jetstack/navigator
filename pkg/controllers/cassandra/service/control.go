package service

import (
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
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

func (e *defaultCassandraClusterServiceControl) Sync(c *v1alpha1.CassandraCluster) error {
	return nil
}
