package seedprovider

import (
	serviceutil "github.com/jetstack/navigator/pkg/controllers/cassandra/service/util"

	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
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
	return serviceutil.SyncService(
		cluster,
		e.kubeClient,
		e.serviceLister,
		ServiceForCluster,
		updateServiceForCluster,
	)
}
