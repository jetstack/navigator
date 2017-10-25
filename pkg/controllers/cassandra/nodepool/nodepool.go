package nodepool

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta2"
	"k8s.io/client-go/tools/record"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterNodepoolControl struct {
	kubeClient        kubernetes.Interface
	statefulsetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultCassandraClusterNodepoolControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulsetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterNodepoolControl{
		kubeClient:        kubeClient,
		statefulsetLister: statefulsetLister,
		recorder:          recorder,
	}
}

func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	for _, pool := range cluster.Spec.NodePools {
		glog.V(4).Infof("syncing nodepool: %#v", pool)
	}
	return nil
}
