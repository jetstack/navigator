package service

import (
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterServiceControl struct {
	kubeClient    kubernetes.Interface
	serviceLister corelisters.ServiceLister
	recorder      record.EventRecorder
}

var _ Interface = &defaultCassandraClusterServiceControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	serviceLister corelisters.ServiceLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterServiceControl{
		kubeClient:    kubeClient,
		serviceLister: serviceLister,
		recorder:      recorder,
	}
}

func (e *defaultCassandraClusterServiceControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	svc := ServiceForCluster(cluster)
	_, err := e.kubeClient.CoreV1().Services(svc.Namespace).Update(svc)
	if k8sErrors.IsNotFound(err) {
		_, err = e.kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
	}
	return err
}
