package elasticsearch

import (
	"fmt"

	apps "k8s.io/api/apps/v1beta1"

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

// Execute will scale the s.NodePool statefulset to the desired number of
// replicas. It will refuse to scale if the cluster is not in a 'state to scale'
// as defined by s.canScaleNodePool.
func (s *Scale) Execute(state *controllers.State) error {
	if s.NodePool == nil || s.Cluster == nil {
		return fmt.Errorf("internal error: nodepool and cluster must be set")
	}

	statefulSetName := util.NodePoolResourceName(s.Cluster, s.NodePool)
	statefulSet, err := state.StatefulSetLister.StatefulSets(s.Cluster.Namespace).Get(statefulSetName)
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

	shouldProceed, err := s.canScaleNodePool(state, statefulSet, replicaDiff)
	if err != nil {
		return err
	}

	if !shouldProceed {
		return fmt.Errorf("cluster not ready to apply scale")
	}

	ssCopy := statefulSet.DeepCopy()
	ssCopy.Spec.Replicas = &s.Replicas
	_, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
	if err != nil {
		return err
	}

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
	allPilots, err := s.pilotsForNodePool(state, s.Cluster.Name, s.NodePool.Name)
	if err != nil {
		return false, err
	}
	toRemove, err := determinePilotsToBeRemoved(allPilots, statefulSet, replicaDiff)
	if err != nil {
		return false, err
	}
	for _, p := range toRemove {
		documentCount := p.Status.Elasticsearch.Documents
		if documentCount == nil ||
			*documentCount > 0 {
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

func (s *Scale) pilotsForNodePool(state *controllers.State, clusterName, poolName string) ([]*v1alpha1.Pilot, error) {
	selector, err := util.SelectorForNodePool(clusterName, poolName)
	if err != nil {
		return nil, err
	}
	return state.PilotLister.Pilots(s.Cluster.Namespace).List(selector)
}
