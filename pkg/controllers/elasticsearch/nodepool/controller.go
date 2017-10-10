package nodepool

import (
	"fmt"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	listersv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

type Interface interface {
	Sync(*v1alpha1.ElasticsearchCluster) error
}

// statefulElasticsearchClusterNodePoolControl manages the lifecycle of a
// stateful node pool. It can be used to create, update and delete pools.
//
// This is an implementation of the ElasticsearchClusterNodePoolControl interface
// as defined in interfaces.go.
type statefulElasticsearchClusterNodePoolControl struct {
	kubeClient        kubernetes.Interface
	navigatorClient   clientset.Interface
	statefulSetLister appslisters.StatefulSetLister
	podLister         corelisters.PodLister
	pilotLister       listersv1alpha1.PilotLister

	recorder record.EventRecorder
}

var _ Interface = &statefulElasticsearchClusterNodePoolControl{}

func NewController(
	kubeClient kubernetes.Interface,
	navigatorClient clientset.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	podLister corelisters.PodLister,
	pilotLister listersv1alpha1.PilotLister,
	recorder record.EventRecorder,
) Interface {
	return &statefulElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		navigatorClient:   navigatorClient,
		statefulSetLister: statefulSetLister,
		podLister:         podLister,
		pilotLister:       pilotLister,
		recorder:          recorder,
	}
}

func (e *statefulElasticsearchClusterNodePoolControl) Sync(c *v1alpha1.ElasticsearchCluster) error {
	if c.Status.NodePools == nil {
		c.Status.NodePools = map[string]v1alpha1.ElasticsearchClusterNodePoolStatus{}
	}
	err := e.reconcileNodePools(c)

	if err != nil {
		return fmt.Errorf("error reconciling node pools: %s", err.Error())
	}

	for _, np := range c.Spec.NodePools {
		npStatus, err := e.syncNodePool(c, &np)
		if err != nil {
			return fmt.Errorf("error syncing nodepool: %s", err.Error())
		}
		c.Status.NodePools[np.Name] = npStatus
	}

	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) syncNodePool(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (v1alpha1.ElasticsearchClusterNodePoolStatus, error) {
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
	npStatus := c.Status.NodePools[np.Name]
	statusCopy := *npStatus.DeepCopy()
	// get the selector for the node pool
	sel, err := util.SelectorForNodePool(c, np)
	if err != nil {
		return statusCopy, fmt.Errorf("error creating label selector for node pool '%s': %s", np.Name, err.Error())
	}
	// list statefulsets matching the selector
	sets, err := e.statefulSetLister.StatefulSets(c.Namespace).List(sel)
	if err != nil {
		return statusCopy, err
	}
	// if more than one statefulset matches the labels, exit here to be safe
	if len(sets) > 1 {
		return statusCopy, fmt.Errorf("multiple StatefulSets match label selector (%s) for node pool '%s'", sel.String(), np.Name)
	}
	expected, err := nodePoolStatefulSet(c, np)
	if err != nil {
		return statusCopy, fmt.Errorf("error generating StatefulSet: %s", err.Error())
	}
	// TODO: extend this to more complex logic than a simple 'create'
	// e.g. queue a new node pool introduced event for Pilots to watch for
	if len(sets) == 0 {
		_, err := e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Create(expected)
		return statusCopy, err
	}
	// this is safe as the above code ensures we only have one element in the array
	actual := sets[0]
	statusCopy.ReadyReplicas = int64(actual.Status.ReadyReplicas)
	if actual.Annotations == nil {
		return statusCopy, fmt.Errorf("StatefulSet '%s' does not contain node pool hash annotation", actual.Name)
	}

	// compare the hashes of the expected and actual node pool
	actualNodePoolHash := actual.Annotations[util.NodePoolHashAnnotationKey]
	if len(actualNodePoolHash) == 0 {
		return statusCopy, fmt.Errorf("StatefulSet '%s' contains empty node pool hash annotation", actual.Name)
	}
	expectedNodePoolHash := expected.Annotations[util.NodePoolHashAnnotationKey]

	// if the node pool hashes do not match, we perform an Update operation on the StatefulSet
	if actualNodePoolHash != expectedNodePoolHash {
		_, err := e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(expected)
		return statusCopy, err
	}

	expectedContainers := expected.Spec.Template.Spec.Containers
	actualContainers := actual.Spec.Template.Spec.Containers
	if len(expectedContainers) != len(actualContainers) {
		_, err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(expected)
		return statusCopy, err
	}
	// check the images are up to date
	for i := 0; i < len(expectedContainers); i++ {
		if expectedContainers[i].Image != actualContainers[i].Image {
			_, err := e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(expected)
			return statusCopy, err
		}
	}

	err = e.syncPilotResources(c, np, actual)
	if err != nil {
		return statusCopy, err
	}

	// the hashes match, which means the properties of the node pool have not changed
	return statusCopy, nil
}

func (e *statefulElasticsearchClusterNodePoolControl) syncPilotResources(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool, ss *appsv1beta1.StatefulSet) error {
	// TODO: use labels to limit which pods we list to save memory
	allPods, err := e.podLister.Pods(c.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	for _, pod := range allPods {
		isMember, err := util.PodControlledByCluster(c, pod, e.statefulSetLister)

		if err != nil {
			return fmt.Errorf("error checking if pod is controller by elasticsearch cluster: %s", err.Error())
		}

		if isMember {
			err = e.ensurePilotResourceExists(c, np, pod)
			if err != nil {
				return fmt.Errorf("error ensuring pilot resource exists for pod '%s': %s", pod.Name, err.Error())
			}
		}
	}
	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) ensurePilotResourceExists(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool, pod *apiv1.Pod) error {
	_, err := e.pilotLister.Pilots(pod.Namespace).Get(pod.Name)
	if apierrors.IsNotFound(err) {
		return e.createPilotResource(c, np, pod)
	}
	if err != nil {
		return err
	}
	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) createPilotResource(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool, pod *apiv1.Pod) error {
	pilot := &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
		},
	}
	_, err := e.navigatorClient.NavigatorV1alpha1().Pilots(pod.Namespace).Create(pilot)
	return err
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
