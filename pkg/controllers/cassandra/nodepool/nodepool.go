package nodepool

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	"github.com/golang/glog"

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

func (e *defaultCassandraClusterNodepoolControl) createStatefulSet(
	cluster *v1alpha1.CassandraCluster,
	nodePool *v1alpha1.CassandraClusterNodePool,
) error {
	desiredSet := StatefulSetForCluster(cluster, nodePool)
	client := e.kubeClient.AppsV1beta1().StatefulSets(cluster.Namespace)
	lister := e.statefulSetLister.StatefulSets(desiredSet.Namespace)
	existingSet, err := lister.Get(desiredSet.Name)
	// StatefulSet already exists
	if err == nil {
		// XXX Temporary hack to enable scale out until we implement ScaleOut action.
		if *existingSet.Spec.Replicas < *desiredSet.Spec.Replicas {
			existingSet = existingSet.DeepCopy()
			existingSet.Spec.Replicas = desiredSet.Spec.Replicas
			_, err = client.Update(existingSet)
			return err
		}
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
	return e.updateStatus(cluster)
}

func (e *defaultCassandraClusterNodepoolControl) updateStatus(cluster *v1alpha1.CassandraCluster) error {
	cluster.Status.NodePools = map[string]v1alpha1.CassandraClusterNodePoolStatus{}
	sets, err := util.StatefulSetsForCluster(cluster, e.statefulSetLister)
	if err != nil {
		return err
	}
	// Create a NodePoolStatus for each statefulset that is controlled by this cluster.
	for ssName, ss := range sets {
		clusterName, npName, err := util.ParseNodePoolResourceName(ssName)
		if err != nil {
			glog.Errorf("Error parsing statefulset name: %s", err)
		}
		if clusterName != cluster.Name {
			glog.Errorf("statefulset name %q did not contain cluster name %q", ssName, cluster.Name)
		}
		nps := cluster.Status.NodePools[npName]
		nps.ReadyReplicas = ss.Status.ReadyReplicas
		cluster.Status.NodePools[npName] = nps
	}
	return nil
}
