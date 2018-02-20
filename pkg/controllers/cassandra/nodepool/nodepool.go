package nodepool

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterNodepoolControl struct {
	kubeClient        kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	pods              corelisters.PodLister
	nodes             corelisters.NodeLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultCassandraClusterNodepoolControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	pods corelisters.PodLister,
	nodes corelisters.NodeLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterNodepoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		pods:              pods,
		nodes:             nodes,
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

// Borrowed from pilot.go, should be moved into a common package?
func (e *defaultCassandraClusterNodepoolControl) clusterPods(cluster *v1alpha1.CassandraCluster) ([]*v1.Pod, error) {
	var clusterPods []*v1.Pod
	allPods, err := e.pods.Pods(cluster.Namespace).List(labels.Everything())
	if err != nil {
		return clusterPods, err
	}
	for _, pod := range allPods {
		podControlledByCluster, err := controllers.PodControlledByCluster(
			cluster,
			pod,
			e.statefulSetLister,
		)
		if err != nil {
			return clusterPods, err
		}
		if !podControlledByCluster {
			continue
		}
		clusterPods = append(clusterPods, pod)
	}
	return clusterPods, nil
}

// Return the lowest int from a slice of ints
func min(v []int) (m int) {
	if len(v) > 0 {
		m = v[0]
	} else {
		panic("min called with empty slice")
	}
	for _, e := range v {
		if e < m {
			m = e
		}
	}
	return
}

func (e *defaultCassandraClusterNodepoolControl) labelSeedNodes(cluster *v1alpha1.CassandraCluster) error {
	pods, err := e.clusterPods(cluster)
	if err != nil {
		return err
	}

	podsInZone := make(map[string][]*v1.Pod)

	// Group pods by failure zone
	for _, pod := range pods {
		fmt.Println("this pod is on", pod.Spec.NodeName)
		node, err := e.nodes.Get(pod.Spec.NodeName)
		if err != nil {
			return err
		}
		fmt.Println("node is in region:", node.Labels["failure-domain.beta.kubernetes.io/zone"])
		nodeRegion := node.Labels["failure-domain.beta.kubernetes.io/zone"]
		podsInZone[nodeRegion] = append(podsInZone[nodeRegion], pod)
	}

	// Choose the lowest StatefulSet member in each failure zone and mark it as a seed
	for _, pods := range podsInZone {
		numberedPods := make(map[int]*v1.Pod)
		ns := []int{}
		for _, p := range pods {
			elements := strings.Split(p.Name, "-")
			n, err := strconv.Atoi(elements[len(elements)-1])
			if err != nil {
				return err
			}
			numberedPods[n] = p
			ns = append(ns, n)
		}

		pod := numberedPods[min(ns)]
		pod.Labels["seed"] = "true"
		e.kubeClient.CoreV1().Pods(pod.Namespace).Update(pod)
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

	err := e.labelSeedNodes(cluster)
	if err != nil {
		return err
	}

	err = e.removeUnusedStatefulSets(cluster)
	return err
}

func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	return e.syncStatefulSets(cluster)
}
