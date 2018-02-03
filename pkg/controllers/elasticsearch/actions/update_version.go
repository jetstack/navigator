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

func (c *UpdateVersion) Message() string {
	if c.podToUpdateName == "" {
		return fmt.Sprintf("Updated node pool %q to version %q", c.NodePool.Name, c.Cluster.Spec.Version.String())
	}
	return fmt.Sprintf("Updated pod %q to version %q", c.podToUpdateName, c.Cluster.Spec.Version.String())
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

	// TODO: ensure shard reallocation is disabled
	if c.Cluster.Status.Health == v1alpha1.ElasticsearchClusterHealthRed {
		err = fmt.Errorf("Cluster is in a red state, refusing to upgrade node pool %q", c.NodePool.Name)
		state.Recorder.Eventf(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}

	currentVersionStr, ok := statefulSet.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation]
	if !ok {
		err = fmt.Errorf("StatefulSet %q does not have an Elasticsearch version annotation", statefulSet.Name)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}
	// attempt to parse the version
	currentVersion, err := semver.NewVersion(currentVersionStr)
	if err != nil {
		err = fmt.Errorf("Invalid version string %q on statefulset %q: %v", currentVersionStr, statefulSet.Name, err)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}
	// this means the statefulset is already up to date. exit early.
	if c.Cluster.Spec.Version.Equal(*currentVersion) {
		return nil
	}

	newImageSpec, err := esImageToUse(&c.Cluster.Spec)
	if err != nil {
		err = fmt.Errorf("Cannot determine Elasticsearch image to use for cluster: %v", err)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}

	desiredReplicasPtr := statefulSet.Spec.Replicas
	if desiredReplicasPtr == nil {
		err = fmt.Errorf("Desired replicas on statefulset %q cannot be nil", statefulSet.Name)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}
	currentPartition := int32(0)
	if statefulSet.Spec.UpdateStrategy.RollingUpdate != nil &&
		statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		currentPartition = *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition
	}
	desiredReplicas := *desiredReplicasPtr
	updatedReplicas := statefulSet.Status.UpdatedReplicas
	currentReplicas := statefulSet.Status.CurrentReplicas
	glog.V(4).Infof("ElasticsearchCluster node pool %q/%q - desired replicas: %d, updated replicas: %d, current replicas: %d, current partition: %d", c.Cluster.Name, c.NodePool.Name, desiredReplicas, updatedReplicas, currentReplicas, currentPartition)

	// we subtract 1 here because we count from 0 when counting replicas
	nextPartitionToUse := int32(desiredReplicas - updatedReplicas - 1)

	// we have already ensured the cluster is not in a red state near the start of the function
	if desiredReplicas != statefulSet.Status.ReadyReplicas {
		// return nil so that this action will be re-queued once ReadyReplicas
		// has been updated
		return nil
	}

	// check if the pod templates image field needs updating
	if len(statefulSet.Spec.Template.Spec.Containers) == 0 {
		err = fmt.Errorf("No containers specified on statefulset %q pod template", statefulSet.Name)
		state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
		return nil
	}
	existingImage := statefulSet.Spec.Template.Spec.Containers[0].Image

	ssCopy := statefulSet.DeepCopy()
	updateMsg := ""
	newImage := newImageSpec.Repository + ":" + newImageSpec.Tag
	switch {
	case c.pilotsUpToDate(state, statefulSet) == nil:
		// we only set this once the upgrade is complete for the node pool
		ssCopy.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation] = c.Cluster.Spec.Version.String()
		updateMsg = fmt.Sprintf("Updated node pool %q to version %q", c.NodePool.Name, c.Cluster.Spec.Version.String())
	case currentPartition != nextPartitionToUse || existingImage != newImage:
		if ssCopy.Spec.UpdateStrategy.RollingUpdate == nil {
			ssCopy.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateStatefulSetStrategy{}
		}
		// ensure previously updated replica has finished updating and is
		// reporting the new version
		lastReplicaUpdated := nextPartitionToUse + 1
		if lastReplicaUpdated < desiredReplicas {
			lastPilotUpdatedName := pilotNameForStatefulSetReplica(ssCopy.Name, lastReplicaUpdated)
			lastPilotUpdated, err := state.PilotLister.Pilots(c.Cluster.Namespace).Get(lastPilotUpdatedName)
			if err != nil {
				return err
			}
			if lastPilotUpdated.Status.Elasticsearch == nil ||
				lastPilotUpdated.Status.Elasticsearch.Version == "" {
				err := fmt.Errorf("Pilot %q has not finished updating to version %q", lastPilotUpdated.Name, c.Cluster.Spec.Version.String())
				state.Recorder.Event(c.Cluster, core.EventTypeWarning, "Err"+c.Name(), err.Error())
				return nil
			}
		}
		// update the statefulset spec
		ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition = &nextPartitionToUse
		ssCopy.Spec.Template.Spec.Containers[0].Image = newImage
		podToUpdateName := pilotNameForStatefulSetReplica(ssCopy.Name, nextPartitionToUse)
		updateMsg = fmt.Sprintf("Updating replica %s to version %s", podToUpdateName, c.Cluster.Spec.Version.String())
	default:
		return nil
	}

	_, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
	if err != nil {
		return err
	}

	state.Recorder.Eventf(c.Cluster, core.EventTypeNormal, c.Name(), updateMsg)

	return nil
}

func (c *UpdateVersion) pilotsUpToDate(state *controllers.State, statefulSet *apps.StatefulSet) error {
	pilots, err := pilotsForStatefulSet(state, c.Cluster, c.NodePool, statefulSet)
	if err != nil {
		return err
	}
	for _, p := range pilots {
		if p.Status.Elasticsearch == nil ||
			p.Status.Elasticsearch.Version == "" {
			return fmt.Errorf("pilot %q version unknown", p.Name)
		}
		if c.Cluster.Spec.Version.String() != p.Status.Elasticsearch.Version {
			return fmt.Errorf("pilot %q is version %q", p.Name, p.Status.Elasticsearch.Version)
		}
	}
	return nil
}
