package actions

import (
	"fmt"

	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerror "k8s.io/apimachinery/pkg/util/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
	utilapi "github.com/jetstack/navigator/pkg/util/api"
)

type Scale struct {
	// The Elasticsearch cluster resource being operated on
	Cluster *v1alpha1.ElasticsearchCluster
	// The node pool being scaled
	NodePool *v1alpha1.ElasticsearchClusterNodePool
	// Number of replicas to scale to
	Replicas int32
}

var _ controllers.Action = &Scale{}

func (c *Scale) Name() string {
	return "Scale"
}

func (c *Scale) Message() string {
	return fmt.Sprintf("Scaled node pool %q to %d replicas", c.NodePool.Name, c.NodePool.Replicas)
}

// Execute will scale the s.NodePool statefulset to the desired number of
// replicas. It will refuse to scale if the cluster is not in a 'state to scale'
// as defined by s.canScaleNodePool.
func (s *Scale) Execute(state *controllers.State) error {
	if s.NodePool == nil || s.Cluster == nil {
		return fmt.Errorf("internal error: nodepool and cluster must be set")
	}

	statefulSetName := util.NodePoolResourceName(s.Cluster, s.NodePool)
	statefulSet, err := state.StatefulSetLister.StatefulSets(s.Cluster.Namespace).Get(statefulSetName)
	if k8sErrors.IsNotFound(err) {
		state.Recorder.Eventf(s.Cluster, core.EventTypeWarning, "Err"+s.Name(), "Cannot get statefulset %q for node pool %q: %v", statefulSetName, s.NodePool.Name, err)
		return nil
	}
	if err != nil {
		return err
	}

	currentReplicas := statefulSet.Spec.Replicas
	// TODO: not sure if we should treat nil as 1 or 0 instead of erroring
	if currentReplicas == nil {
		return fmt.Errorf("current number of replicas on statefulset cannot be nil")
	}
	replicaDiff := s.Replicas - (*currentReplicas)
	// if this scale won't change the number of replicas, we can return early
	if replicaDiff == 0 {
		return nil
	}

	canScale, err := s.canScaleNodePool(state, statefulSet, replicaDiff)
	if err != nil {
		return err
	}
	if !canScale {
		return nil
	}

	ssCopy := statefulSet.DeepCopy()
	ssCopy.Spec.Replicas = &s.Replicas
	_, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
	if err != nil {
		return err
	}
	state.Recorder.Eventf(s.Cluster, core.EventTypeNormal, s.Name(), "Scaled node pool %q to %d replicas", s.NodePool.Name, s.NodePool.Replicas)

	return nil
}

// canScaleNodePool will determine whether this scale operation is 'valid'.
// Valid is defined as:
// - if the node pool does not have a 'data' node pool role, a scale is always
//   valid.
// - if the scale is not a 'scale down', it is always valid.
// - if all pilots effected by the scale (e.g. those that would be removed)
//   have been drained of all shards, we can scale down.
// - otherwise reject the scale down.
func (s *Scale) canScaleNodePool(state *controllers.State, statefulSet *apps.StatefulSet, replicaDiff int32) (bool, error) {
	// always allow scale up, or staying the same (no-op)
	if replicaDiff >= 0 {
		return true, nil
	}
	// we can always scale down non-data roles, as validation should
	// ensure that at least one data, ingest and master node exists already
	if !utilapi.ContainsElasticsearchRole(s.NodePool.Roles, v1alpha1.ElasticsearchRoleData) {
		return true, nil
	}

	// outline of what goes on here:
	// - determine if we were to scale down, which Pods/Pilots would be removed
	//   from the cluster
	// - if excludeShards is false on those pilots, return false
	// - if documentCount is >0 on those pilots, return false
	// - otherwise return true
	allPilots, err := pilotsForStatefulSet(state, s.Cluster, s.NodePool, statefulSet)
	if k8sErrors.IsNotFound(err) {
		state.Recorder.Eventf(s.Cluster, core.EventTypeWarning, "Err"+s.Name(), "Could not get all pilots for statefulset %q: %v", statefulSet.Name, err)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	toRemove, err := determinePilotsToBeRemoved(allPilots, statefulSet, replicaDiff)
	if err != nil {
		return false, err
	}
	for _, p := range toRemove {
		if p.Status.Elasticsearch == nil ||
			p.Status.Elasticsearch.Documents == nil {
			state.Recorder.Eventf(s.Cluster, core.EventTypeWarning, "Err"+s.Name(), "Could not determine document count for pilot %q", p.Name)
			return false, nil
		}
		documentCount := *p.Status.Elasticsearch.Documents
		if documentCount > 0 {
			state.Recorder.Eventf(s.Cluster, core.EventTypeNormal, "Err"+s.Name(), "Pilot %q still contains %d documents. Will not scale down until empty.", p.Name, documentCount)
			return false, nil
		}
	}
	return true, nil
}

// determinePilotsToBeRemoved will return which pilots would be removed after
// this scale operation.
// - allPilots is a slice of all of the pilots that should be considered in the
//  calculation, and should generally be a list of all current pilots in the
//  statefulset.
// - statefulSet is the statefulset being scaled
// - replicaDiff is the how many replicas are being added to the statefulset.
//   If greater than zero, an empty list and no error is returned.
func determinePilotsToBeRemoved(allPilots []*v1alpha1.Pilot, statefulSet *apps.StatefulSet, replicaDiff int32) ([]*v1alpha1.Pilot, error) {
	if replicaDiff >= 0 {
		return nil, nil
	}
	currentReplicas := statefulSet.Spec.Replicas
	// TODO: not sure if we should treat nil as 1 or 0 instead of erroring
	if currentReplicas == nil {
		return nil, fmt.Errorf("current replicas of statefulset cannot be nil")
	}
	var toBeRemoved []*v1alpha1.Pilot
	desiredReplicas := *currentReplicas + replicaDiff
	for i := (*currentReplicas) - 1; i >= desiredReplicas; i-- {
		pilotName := pilotNameForStatefulSetReplica(statefulSet.Name, i)
		for _, p := range allPilots {
			if pilotName == p.Name && statefulSet.Namespace == p.Namespace {
				toBeRemoved = append(toBeRemoved, p)
				break
			}
		}
	}
	return toBeRemoved, nil
}

func pilotNameForStatefulSetReplica(setName string, replica int32) string {
	return fmt.Sprintf("%s-%d", setName, replica)
}

func pilotsForStatefulSet(state *controllers.State, cluster *v1alpha1.ElasticsearchCluster, nodePool *v1alpha1.ElasticsearchClusterNodePool, statefulSet *apps.StatefulSet) ([]*v1alpha1.Pilot, error) {
	replicasPtr := statefulSet.Spec.Replicas
	if replicasPtr == nil {
		return nil, fmt.Errorf("statefulset.spec.replicas cannot be nil")
	}
	replicas := *replicasPtr
	// TODO: read the cluster name and node pool name from the statefulset
	// metadata instead of the Scale structure so we can make this a package
	// function. For now, this way is safest until we are sure these
	// labels are going to remain stable
	selector, err := util.SelectorForNodePool(cluster.Name, nodePool.Name)
	if err != nil {
		return nil, err
	}
	pilots, err := state.PilotLister.Pilots(cluster.Namespace).List(selector)
	if err != nil {
		return nil, err
	}
	var errs []error
	var output []*v1alpha1.Pilot
Outer:
	for i := int32(0); i < replicas; i++ {
		pilotName := pilotNameForStatefulSetReplica(statefulSet.Name, i)
		for _, p := range pilots {
			if p.Name == pilotName {
				output = append(output, p)
				continue Outer
			}
		}
		errs = append(errs, fmt.Errorf("pilot %q not found", pilotName))
	}
	if len(errs) > 0 {
		return nil, &k8sErrors.StatusError{
			ErrStatus: metav1.Status{
				Message: utilerror.NewAggregate(errs).Error(),
				Reason:  metav1.StatusReasonNotFound,
			},
		}
	}
	return output, nil
}
