package service

import (
	"k8s.io/client-go/kubernetes"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterServiceControl struct {
	kubeClient kubernetes.Interface
}

var _ Interface = &defaultCassandraClusterServiceControl{}

func NewControl(
	kubeClient kubernetes.Interface,
) Interface {
	return &defaultCassandraClusterServiceControl{
		kubeClient: kubeClient,
	}
}

func (e *defaultCassandraClusterServiceControl) Sync(c *v1alpha1.CassandraCluster) error {
	return nil
}
