package nodepool

import (
	"fmt"

	"github.com/golang/glog"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
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
	pods              corelisters.PodLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultCassandraClusterNodepoolControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	pods corelisters.PodLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterNodepoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		pods:              pods,
		recorder:          recorder,
	}
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
	lister := e.statefulSetLister.StatefulSets(cluster.Namespace)
	selector, err := util.SelectorForCluster(cluster)
	if err != nil {
		return err
	}
	existingSets, err := lister.List(selector)
	if err != nil {
		return err
	}
	for _, set := range existingSets {
		err := util.OwnerCheck(set, cluster)
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

func (e *defaultCassandraClusterNodepoolControl) labelSeedNodes(
	cluster *v1alpha1.CassandraCluster,
	set *appsv1beta1.StatefulSet,
) error {
	// TODO: make number of seed nodes configurable
	pod, err := e.pods.Pods(cluster.Namespace).Get(fmt.Sprintf("%s-%d", set.Name, 0))
	if err != nil {
		glog.Warningf("Couldn't get stateful set pod: %v", err)
		return nil
	}

	// only label if the current label is incorrect
	if pod.Labels["seed"] != "true" {
		podCopy := pod.DeepCopy()
		podCopy.Labels["seed"] = "true"
		_, err := e.kubeClient.CoreV1().Pods(podCopy.Namespace).Update(podCopy)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *defaultCassandraClusterNodepoolControl) createOrUpdateStatefulSet(
	cluster *v1alpha1.CassandraCluster,
	nodePool *v1alpha1.CassandraClusterNodePool,
) error {
	desiredSet := StatefulSetForCluster(cluster, nodePool)
	client := e.kubeClient.AppsV1beta1().StatefulSets(cluster.Namespace)
	lister := e.statefulSetLister.StatefulSets(desiredSet.Namespace)
	existingSet, err := lister.Get(desiredSet.Name)
	if k8sErrors.IsNotFound(err) {
		_, err = client.Create(desiredSet)
		return err
	}
	if err != nil {
		return err
	}
	err = util.OwnerCheck(existingSet, cluster)
	if err != nil {
		return err
	}

	err = e.labelSeedNodes(cluster, existingSet)
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
