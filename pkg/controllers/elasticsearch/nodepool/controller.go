package nodepool

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

type Interface interface {
	Sync(*v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error)
}

// statefulElasticsearchClusterNodePoolControl manages the lifecycle of a
// stateful node pool. It can be used to create, update and delete pools.
//
// This is an implementation of the ElasticsearchClusterNodePoolControl interface
// as defined in interfaces.go.
type statefulElasticsearchClusterNodePoolControl struct {
	kubeClient        kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister

	recorder record.EventRecorder
}

var _ Interface = &statefulElasticsearchClusterNodePoolControl{}

func NewController(
	kubeClient kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) Interface {
	return &statefulElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
	}
}

func (e *statefulElasticsearchClusterNodePoolControl) Sync(c *v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error) {
	err := e.reconcileNodePools(c)

	if err != nil {
		return c.Status, fmt.Errorf("error reconciling node pools: %s", err.Error())
	}

	for _, np := range c.Spec.NodePools {
		err := e.syncNodePool(c, &np)
		if err != nil {
			return c.Status, fmt.Errorf("error syncing nodepool: %s", err.Error())
		}
	}

	return c.Status, nil
}

func (e *statefulElasticsearchClusterNodePoolControl) syncNodePool(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) error {
	// lookup existing StatefulSet with appropriate labels for np in cluster c
	// if multiple exist, exit with an error
	// if one exists:
	//		- generate the expected StatefulSet manifest
	// 		- compare the expected hashes
	//		- if they differ, we perform an update of the StatefulSet
	//		- otherwise, we check if any additional fields (e.g. image)
	//		  have changed
	// if none exist:
	//		- generate the expected StatefulSet manifest
	//		- create the StatefulSet

	// get the selector for the node pool
	sel, err := util.SelectorForNodePool(c, np)
	if err != nil {
		return fmt.Errorf("error creating label selector for node pool '%s': %s", np.Name, err.Error())
	}
	// list statefulsets matching the selector
	sets, err := e.statefulSetLister.StatefulSets(c.Namespace).List(sel)
	if err != nil {
		return err
	}
	// if more than one statefulset matches the labels, exit here to be safe
	if len(sets) > 1 {
		return fmt.Errorf("multiple StatefulSets match label selector (%s) for node pool '%s'", sel.String(), np.Name)
	}
	expected, err := nodePoolStatefulSet(c, np)
	if err != nil {
		return fmt.Errorf("error generating StatefulSet: %s", err.Error())
	}
	// TODO: extend this to more complex logic than a simple 'create'
	// e.g. queue a new node pool introduced event for Pilots to watch for
	if len(sets) == 0 {
		_, err := e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Create(expected)
		return err
	}
	// this is safe as the above code ensures we only have one element in the array
	actual := sets[0]
	if actual.Annotations == nil {
		return fmt.Errorf("StatefulSet '%s' does not contain node pool hash annotation", actual.Name)
	}

	// compare the hashes of the expected and actual node pool
	actualNodePoolHash := actual.Annotations[util.NodePoolHashAnnotationKey]
	if len(actualNodePoolHash) == 0 {
		return fmt.Errorf("StatefulSet '%s' contains empty node pool hash annotation", actual.Name)
	}
	expectedNodePoolHash := expected.Annotations[util.NodePoolHashAnnotationKey]

	// if the node pool hashes do not match, we perform an Update operation on the StatefulSet
	if actualNodePoolHash != expectedNodePoolHash {
		_, err := e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(expected)
		return err
	}

	expectedContainers := expected.Spec.Template.Spec.Containers
	actualContainers := actual.Spec.Template.Spec.Containers
	if len(expectedContainers) != len(actualContainers) {
		_, err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(expected)
		return err
	}
	// check the images are up to date
	for i := 0; i < len(expectedContainers); i++ {
		if expectedContainers[i].Image != actualContainers[i].Image {
			_, err := e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(expected)
			return err
		}
	}

	// the hashes match, which means the properties of the node pool have not changed
	return nil
}

// reconcileNodePools will look up all node pools that are owned by this
// ElasticsearchCluster resource, and delete any that are no longer referenced.
// This is used to delete old node pools that no longer exist in the cluster
// specification.
func (e *statefulElasticsearchClusterNodePoolControl) reconcileNodePools(c *v1alpha1.ElasticsearchCluster) error {
	// list all statefulsets that match the clusters selector
	// loop through each node pool in c
	sel, err := util.SelectorForCluster(c)
	if err != nil {
		return fmt.Errorf("error creating label selector for cluster '%s': %s", c.Name, err.Error())
	}
	sets, err := e.statefulSetLister.StatefulSets(c.Namespace).List(sel)
	if err != nil {
		return err
	}
	// we delete each statefulset that has the node pool name set to the name
	// of a valid node pool for sets
	for _, np := range c.Spec.NodePools {
		for i, ss := range sets {

			if ss.Labels != nil && ss.Labels[util.NodePoolNameLabelKey] == np.Name {
				sets = append(sets[:i], sets[i+1:]...)
				break
			}
		}
	}
	// delete remaining statefulsets in sets
	for _, ss := range sets {
		err := e.kubeClient.AppsV1beta1().StatefulSets(ss.Namespace).Delete(ss.Name, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *statefulElasticsearchClusterNodePoolControl) recordNodePoolEvent(verb string, cluster v1alpha1.ElasticsearchCluster, pool v1alpha1.ElasticsearchClusterNodePool, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s StatefulNodePool %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), pool.Name, cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s StatefulNodePool %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), pool.Name, cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
