package nodepool

import (
	"fmt"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	apps "k8s.io/api/apps/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
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

func ownerCheck(
	set *apps.StatefulSet,
	cluster *v1alpha1.CassandraCluster,
) error {
	if !metav1.IsControlledBy(set, cluster) {
		ownerRef := metav1.GetControllerOf(set)
		return fmt.Errorf(
			"Foreign owned StatefulSet: "+
				"A StatefulSet with name '%s/%s' already exists, "+
				"but it is controlled by '%v', not '%s/%s'.",
			set.Namespace, set.Name, ownerRef,
			cluster.Namespace, cluster.Name,
		)
	}
	return nil
}

func (e *defaultCassandraClusterNodepoolControl) removeUnusedStatefulSets(
	cluster *v1alpha1.CassandraCluster,
) error {
	expectedStatefulSetNames := map[string]bool{}
	for _, pool := range cluster.Spec.NodePools {
		name := util.NodePoolResourceName(cluster, &pool)
		expectedStatefulSetNames[name] = true
	}
	client := e.kubeClient.AppsV1beta1().StatefulSets(cluster.Namespace)
	selector, err := util.SelectorForCluster(cluster)
	if err != nil {
		return err
	}
	existingSets, err := e.statefulsetLister.
		StatefulSets(cluster.Namespace).
		List(selector)
	if err != nil {
		return err
	}
	for _, set := range existingSets {
		err := ownerCheck(set, cluster)
		if err != nil {
			return err
		}
		_, found := expectedStatefulSetNames[set.Name]
		if !found {
			err := client.Delete(set.Name, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *defaultCassandraClusterNodepoolControl) createOrUpdateStatefulSet(
	cluster *v1alpha1.CassandraCluster,
	nodePool *v1alpha1.CassandraClusterNodePool,
) error {
	client := e.kubeClient.AppsV1beta1().StatefulSets(cluster.Namespace)
	desiredSet := StatefulSetForCluster(cluster, nodePool)
	existingSet, err := e.statefulsetLister.
		StatefulSets(desiredSet.Namespace).
		Get(desiredSet.Name)
	if k8sErrors.IsNotFound(err) {
		_, err = client.Create(desiredSet)
		return err
	}
	if err != nil {
		return err
	}
	err = ownerCheck(existingSet, cluster)
	if err != nil {
		return err
	}
	_, err = client.Update(desiredSet)
	return err
}

func (e *defaultCassandraClusterNodepoolControl) syncStatefulSets(
	cluster *v1alpha1.CassandraCluster,
) error {
	for _, pool := range cluster.Spec.NodePools {
		err := e.createOrUpdateStatefulSet(cluster, &pool)
		if err != nil {
			return err
		}
	}
	err := e.removeUnusedStatefulSets(cluster)
	return err
}

func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	return e.syncStatefulSets(cluster)
}
