package nodepool

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
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

func (e *defaultCassandraClusterNodepoolControl) createStatefulSet(
	cluster *v1alpha1.CassandraCluster,
	nodePool *v1alpha1.CassandraClusterNodePool,
) error {
	desiredSet := StatefulSetForCluster(cluster, nodePool)
	client := e.kubeClient.AppsV1beta1().StatefulSets(cluster.Namespace)
	lister := e.statefulSetLister.StatefulSets(desiredSet.Namespace)
	_, err := lister.Get(desiredSet.Name)
	// StatefulSet already exists
	if err == nil {
		return nil
	}
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	_, err = client.Create(desiredSet)
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	for _, pool := range cluster.Spec.NodePools {
		err := e.createStatefulSet(cluster, &pool)
		if err != nil {
			return err
		}
	}
	return nil
}
