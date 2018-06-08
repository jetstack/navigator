package nodepool

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterNodepoolControl struct {
	kubeClient        kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultCassandraClusterNodepoolControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterNodepoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
	}
}

func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	return e.updateStatus(cluster)
}

func (e *defaultCassandraClusterNodepoolControl) updateStatus(cluster *v1alpha1.CassandraCluster) error {
	cluster.Status.NodePools = map[string]v1alpha1.CassandraClusterNodePoolStatus{}
	sets, err := util.StatefulSetsForCluster(cluster, e.statefulSetLister)
	if err != nil {
		return err
	}
	// Create a NodePoolStatus for each statefulset that is controlled by this cluster.
	for _, ss := range sets {
		npName := ss.Labels[v1alpha1.CassandraNodePoolNameLabel]
		nps := cluster.Status.NodePools[npName]
		nps.ReadyReplicas = ss.Status.ReadyReplicas

		var container *apiv1.Container
		for i, _ := range ss.Spec.Template.Spec.Containers {
			if ss.Spec.Template.Spec.Containers[i].Name == "cassandra" {
				container = &ss.Spec.Template.Spec.Containers[i]
			}
		}

		if container != nil {
			nps.Resources = container.Resources
		}

		cluster.Status.NodePools[npName] = nps
	}
	return nil
}
