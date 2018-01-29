package nodepool

import (
	"fmt"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

type Interface interface {
	Sync(*v1alpha1.ElasticsearchCluster) error
}

// statefulControl manages the lifecycle of a
// stateful node pool. It can be used to create, update and delete pools.
//
// This is an implementation of the ElasticsearchClusterNodePoolControl interface
// as defined in interfaces.go.
type statefulControl struct {
	kubeClient        kubernetes.Interface
	navigatorClient   clientset.Interface
	statefulSetLister appslisters.StatefulSetLister
	podLister         corelisters.PodLister
	pilotLister       listersv1alpha1.PilotLister

	recorder record.EventRecorder
}

var _ Interface = &statefulControl{}

func NewController(
	kubeClient kubernetes.Interface,
	navigatorClient clientset.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	podLister corelisters.PodLister,
	pilotLister listersv1alpha1.PilotLister,
	recorder record.EventRecorder,
) Interface {
	return &statefulControl{
		kubeClient:        kubeClient,
		navigatorClient:   navigatorClient,
		statefulSetLister: statefulSetLister,
		podLister:         podLister,
		pilotLister:       pilotLister,
		recorder:          recorder,
	}
}

func (e *statefulControl) Sync(c *v1alpha1.ElasticsearchCluster) error {
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

func (e *statefulControl) syncNodePool(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (v1alpha1.ElasticsearchClusterNodePoolStatus, error) {
	// lookup existing StatefulSet with appropriate labels for np in cluster c
	// if multiple exist, exit with an error
	// if one exists:
	//		- generate the expected StatefulSet manifest
	//		- compare the expected hashes
	//		- if they differ, we perform an update of the StatefulSet
	//		- otherwise, we check if any additional fields (e.g. image)
	//		  have changed
	// if none exist:
	//		- generate the expected StatefulSet manifest
	//		- create the StatefulSet
	npStatus := c.Status.NodePools[np.Name]
	statusCopy := *npStatus.DeepCopy()

	desiredStatefulSet, err := NodePoolStatefulSet(c, np)
	if err != nil {
		return statusCopy, fmt.Errorf("error generating StatefulSet: %s", err.Error())
	}

	existingStatefulSet, err := e.existingStatefulSet(c, np)
	if apierrors.IsNotFound(err) {
		existingStatefulSet, err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Create(desiredStatefulSet)
		if err != nil {
			return statusCopy, fmt.Errorf("error creating StatefulSet: %s", err.Error())
		}
	} else if err != nil {
		// TODO: log event
		return statusCopy, err
	}

	var hash string
	var ok bool
	if hash, ok = existingStatefulSet.Annotations[v1alpha1.ElasticsearchNodePoolHashAnnotation]; ok {
		// TODO: set collisionCount properly
		desiredHash := util.ComputeNodePoolHash(c, np, util.Int32Ptr(0))
		if desiredHash != hash {
			existingStatefulSet, err = e.updateStatefulSet(desiredHash, existingStatefulSet, desiredStatefulSet)
			if err != nil {
				// TODO: log event to say update of statefulset failed
				return statusCopy, fmt.Errorf("error updating StatefulSet: %s", err.Error())
			}
		}
	}

	statusCopy.ReadyReplicas = int64(existingStatefulSet.Status.ReadyReplicas)

	// the hashes match, which means the properties of the node pool have not changed
	return statusCopy, nil
}

func (e *statefulControl) updateStatefulSet(hash string, existing, new *appsv1beta1.StatefulSet) (*appsv1beta1.StatefulSet, error) {
	copy := existing.DeepCopy()
	copy.Labels = new.Labels
	copy.Annotations = new.Annotations
	copy.Spec.UpdateStrategy = new.Spec.UpdateStrategy
	copy.Spec.Replicas = new.Spec.Replicas
	copy.Spec.PodManagementPolicy = new.Spec.PodManagementPolicy
	copy.Spec.RevisionHistoryLimit = new.Spec.RevisionHistoryLimit
	copy.Spec.Template = new.Spec.Template
	return e.kubeClient.AppsV1beta1().StatefulSets(copy.Namespace).Update(copy)
}

func (e *statefulControl) existingStatefulSet(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (*appsv1beta1.StatefulSet, error) { // get the selector for the node pool
	sel, err := util.SelectorForNodePool(c.Name, np.Name)
	if err != nil {
		return nil, fmt.Errorf("error creating label selector for node pool '%s': %s", np.Name, err.Error())
	}
	// list statefulsets matching the selector
	sets, err := e.statefulSetLister.StatefulSets(c.Namespace).List(sel)
	if err != nil {
		return nil, err
	}
	// if more than one statefulset matches the labels, exit here to be safe
	if len(sets) > 1 {
		return nil, fmt.Errorf("multiple StatefulSets match label selector (%s) for node pool '%s'", sel.String(), np.Name)
	}
	if len(sets) == 0 {
		return nil, apierrors.NewNotFound(schema.GroupResource{}, fmt.Sprintf("statefulset for node pool %q not found", np.Name))
	}
	// this is safe as the above code ensures we only have one element in the array
	return sets[0], nil
}

// reconcileNodePools will look up all node pools that are owned by this
// ElasticsearchCluster resource, and delete any that are no longer referenced.
// This is used to delete old node pools that no longer exist in the cluster
// specification.
func (e *statefulControl) reconcileNodePools(c *v1alpha1.ElasticsearchCluster) error {
	// list all statefulsets that match the clusters selector
	// loop through each node pool in c
	sel, err := util.SelectorForCluster(c.Name)
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

			if ss.Labels != nil && ss.Labels[v1alpha1.ElasticsearchNodePoolHashAnnotation] == np.Name {
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
