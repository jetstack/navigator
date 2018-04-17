package actions

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/golang/glog"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

var emptyVersion = semver.Version{}

type UpdateVersion struct {
	// The Elasticsearch cluster resource being operated on
	Cluster *v1alpha1.ElasticsearchCluster
	// The node pool being scaled
	NodePool *v1alpha1.ElasticsearchClusterNodePool
	// this is used internally to generate the Message()
	podToUpdateName string
}

var _ controllers.Action = &UpdateVersion{}

func (c *UpdateVersion) Name() string {
	return "UpdateVersion"
}

func (c *UpdateVersion) Execute(state *controllers.State) error {
	if c.NodePool == nil || c.Cluster == nil {
		return fmt.Errorf("internal error: nodepool and cluster must be set")
	}

	statefulSetName := util.NodePoolResourceName(c.Cluster, c.NodePool)
	statefulSet, err := state.StatefulSetLister.StatefulSets(c.Cluster.Namespace).Get(statefulSetName)
	if err != nil {
		return err
	}

	// Check the health of the Cluster. If it is Red, we do not proceed.
	if c.Cluster.Status.Health == nil {
		return fmt.Errorf("Could not determine current health of Elasticsearch cluster")
	}

	if *c.Cluster.Status.Health == v1alpha1.ElasticsearchClusterHealthRed {
		err = fmt.Errorf("Cluster is in a red state, refusing to upgrade node pool %q", c.NodePool.Name)
		state.Recorder.Eventf(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}

	// TODO: ensure shard reallocation is disabled in the Elasticsearch cluster

	// Determine the current version of the StatefulSet by looking at the version annotation.
	// This is only updated after the upgrade is completed.
	currentVersionStr, ok := statefulSet.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation]
	if !ok {
		err = fmt.Errorf("StatefulSet %q does not have an Elasticsearch version annotation", statefulSet.Name)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}
	// Attempt to parse the version as semver
	currentVersion, err := semver.NewVersion(currentVersionStr)
	if err != nil {
		err = fmt.Errorf("Invalid version string %q on statefulset %q: %v", currentVersionStr, statefulSet.Name, err)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}

	// If the desired version is equal to the current version as specified in the annotation,
	// then we do not need to update the cluster any further.
	// TODO: what happens if a user requests an upgrade, and half way through the upgrade
	// switches the version number back? Despite half of the replicas being the new version,
	// we would not roll back as the version strings are equal.
	if c.Cluster.Spec.Version.Equal(*currentVersion) {
		return nil
	}

	// Determine the desired 'new' Elasticsearch image to use based on the cluster's spec
	newImageSpec, err := esImageToUse(&c.Cluster.Spec)
	if err != nil {
		err = fmt.Errorf("Cannot determine Elasticsearch image to use for cluster: %v", err)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}

	// Get the current update 'partition'. We use this partition field to control roll outs
	// of the new image version.
	// If the partition is not currently set, it defaults to 0, which means all replicas
	// will be updated at once.
	currentPartition := int32(0)
	if statefulSet.Spec.UpdateStrategy.RollingUpdate != nil &&
		statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		currentPartition = *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition
	}

	// Get the current number of desired replicas from the StatefulSet
	desiredReplicas := *statefulSet.Spec.Replicas
	// Get the current number of 'up to date' replicas (i.e. those currently running the
	// 'desired' new version)
	updatedReplicas := statefulSet.Status.UpdatedReplicas
	// Get the current number of 'old' replicas (i.e. those running the old version)
	currentReplicas := statefulSet.Status.CurrentReplicas
	glog.V(4).Infof("ElasticsearchCluster node pool %q/%q - desired replicas: %d, updated replicas: %d, current replicas: %d, current partition: %d", c.Cluster.Name, c.NodePool.Name, desiredReplicas, updatedReplicas, currentReplicas, currentPartition)

	// Determine what the next partition should be by subtracting the 'updated' replicas
	// (i.e. those already running the new version) from the total replicas desired.
	// Because the 'partition' field on a StatefulSet counts from 0, we also subtract 1
	// from this value.
	nextPartitionToUse := int32(desiredReplicas - updatedReplicas - 1)
	// Because we have subtracted 1 from the value, if desiredReplicas=updatedReplicas (i.e.
	// the update is complete), it is possible the value is <0. If so, we reset it to 0.
	if nextPartitionToUse == -1 {
		nextPartitionToUse = 0
	}

	// If any of the pods in the StatefulSet are not currently 'Ready', we return nil
	// here to allow more time for the pods to become Ready. We could alternatively return
	// an error, but that would cause the item to be immediately requeued. Insead we return nil
	// so that the cluster will be resynced once the pod updates to become Ready (as navigator-controller)
	// watches for changes to Pod resources.
	if desiredReplicas != statefulSet.Status.ReadyReplicas {
		return nil
	}

	// Determine the existing image information on the StatefulSet.
	// If the upgrade is already in progress, this will already be set
	// to the new image.
	if len(statefulSet.Spec.Template.Spec.Containers) == 0 {
		err = fmt.Errorf("No containers specified on statefulset %q pod template", statefulSet.Name)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}
	existingImage := statefulSet.Spec.Template.Spec.Containers[0].Image

	// Create a copy of the StatefulSet in preparation for an update
	ssCopy := statefulSet.DeepCopy()
	updateMsg := ""
	// Construct new image string
	newImage := newImageSpec.Repository + ":" + newImageSpec.Tag

	switch {

	// If every Pilot in the cluster is currently reporting their version as the same as
	// the desired version on the ElasticsearchCluster resource (i.e. all pilots are 'up to date'),
	// we update the version annotation on the StatefulSet.
	case c.pilotsUpToDate(state, statefulSet) == nil:
		// we only set this once the upgrade is complete for the node pool
		ssCopy.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation] = c.Cluster.Spec.Version.String()
		updateMsg = fmt.Sprintf("Updated node pool %q to version %q", c.NodePool.Name, c.Cluster.Spec.Version.String())

	// If the 'nextPartitionToUse' parameter is not equal to the 'currentPartition',
	// **or** if the existing image is not the same as the desired image, we should
	// update the StatefulSet.
	case currentPartition != nextPartitionToUse || existingImage != newImage:
		// Initialise the RollingUpdate field if it's not already (as it is a pointer)
		if ssCopy.Spec.UpdateStrategy.RollingUpdate == nil {
			ssCopy.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateStatefulSetStrategy{}
		}

		// The last replica that was updated is the 'nextPartitionToUse + 1' (i.e. the value of the last partition)
		lastReplicaUpdated := nextPartitionToUse + 1
		// If the lastReplicaUpdated is greater than or equal to the desired replicas, we know the this is the
		// first replica being updated as part of the set.
		// Otherwrise, we need to check to ensure the last Pilot that was updated has *finished* updating (i.e. that it
		// is reporting the correct new version number on the Pilot resource).
		if lastReplicaUpdated < desiredReplicas {
			// Determine the name of the previous Pilot that was updated
			lastPilotUpdatedName := pilotNameForStatefulSetReplica(ssCopy.Name, lastReplicaUpdated)
			// Get a copy of the lastPilotUpdated
			lastPilotUpdated, err := state.PilotLister.Pilots(c.Cluster.Namespace).Get(lastPilotUpdatedName)
			if err != nil {
				return err
			}
			// Ensure the last pilot updated is up to date
			if !pilotVersionUpToDate(c.Cluster.Spec, lastPilotUpdated.Status) {
				err := fmt.Errorf("Pilot %q has not finished updating to version %q", lastPilotUpdated.Name, c.Cluster.Spec.Version.String())
				state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
				return nil
			}
		}
		// Update the StatefulSet spec with the new partition to use
		ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition = &nextPartitionToUse
		// Update the image spec to the new image, if it isn't already
		ssCopy.Spec.Template.Spec.Containers[0].Image = newImage
		// Determine which pilot is going to be updated next and log a message accordingly
		podToUpdateName := pilotNameForStatefulSetReplica(ssCopy.Name, nextPartitionToUse)
		updateMsg = fmt.Sprintf("Updating replica %s to version %s", podToUpdateName, c.Cluster.Spec.Version.String())

	// If this default block is hit, it means the Pilots haven't finished updating (as pilotsUpToDate returned false),
	// however the statefulset *has* been fully updated (e.g. the new image version is set, and the partition is set to 0)
	// meaning the final replica has been told to update.
	// We return nil so that this Action will be fired again once the Pilot posts its desired new version.
	default:
		return nil

	}

	// Perform the updates to the StatefulSet from the above switch block.
	_, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
	if err != nil {
		return err
	}

	// Log an informational message to the user
	state.Recorder.Eventf(c.Cluster, core.EventTypeNormal, c.Name(), updateMsg)

	return nil
}

func (c *UpdateVersion) pilotsUpToDate(state *controllers.State, statefulSet *apps.StatefulSet) error {
	pilots, err := pilotsForStatefulSet(state, c.Cluster, c.NodePool, statefulSet)
	if err != nil {
		return err
	}
	for _, p := range pilots {
		if !pilotVersionUpToDate(c.Cluster.Spec, p.Status) {
			return fmt.Errorf("Pilot %q is not up to date", p.Name)
		}
	}
	return nil
}

func pilotVersionUpToDate(c v1alpha1.ElasticsearchClusterSpec, p v1alpha1.PilotStatus) bool {
	if p.Elasticsearch == nil {
		return false
	}
	if p.Elasticsearch.Version == nil {
		return false
	}
	if !p.Elasticsearch.Version.Equal(c.Version) {
		return false
	}
	return true
}
