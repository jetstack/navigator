package actions

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/golang/glog"
	apps "k8s.io/api/apps/v1beta1"

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
		return fmt.Errorf("cluster is in a red state, refusing to upgrade node pool %q", c.NodePool.Name)
	}

	currentVersionStr, ok := statefulSet.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation]
	if !ok {
		return fmt.Errorf("cannot determine existing Elasticsearch version of statefulset %q", statefulSet.Name)
	}
	// attempt to parse the version
	currentVersion, err := semver.NewVersion(currentVersionStr)
	if err != nil {
		return fmt.Errorf("error parsing existing Elasticsearch version: %v", err)
	}
	// this means the statefulset is already up to date. exit early.
	if c.Cluster.Spec.Version.Equal(*currentVersion) {
		return nil
	}

	newImageSpec, err := esImageToUse(&c.Cluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot determine new image details to use: %v", err)
	}

	desiredReplicasPtr := statefulSet.Spec.Replicas
	if desiredReplicasPtr == nil {
		return fmt.Errorf("desired replicas on statefulset cannot be nil")
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
		return fmt.Errorf("all replicas in node pool must be ready before updating")
	}

	// check if the pod templates image field needs updating
	if len(statefulSet.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("internal error: no containers specified on statefulset pod template")
	}

	existingImage := statefulSet.Spec.Template.Spec.Containers[0].Image
	newImage := newImageSpec.Repository + ":" + newImageSpec.Tag
	ssCopy := statefulSet.DeepCopy()
	if existingImage != newImage {
		if ssCopy.Spec.UpdateStrategy.RollingUpdate == nil {
			ssCopy.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateStatefulSetStrategy{}
		}
		// we update the image, but don't allow any pods to be updated yet
		ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition = ssCopy.Spec.Replicas
		ssCopy.Spec.Template.Spec.Containers[0].Image = newImage
	}

	// we are ready to reduce the partition number and begin updating the next
	// replica
	if currentPartition != nextPartitionToUse && nextPartitionToUse >= 0 {
		if ssCopy.Spec.UpdateStrategy.RollingUpdate == nil {
			ssCopy.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateStatefulSetStrategy{}
		}
		ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition = &nextPartitionToUse
	}

	if c.pilotsUpToDate(state, statefulSet) == nil {
		// we only set this once the upgrade is complete for the node pool
		ssCopy.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation] = c.Cluster.Spec.Version.String()
		// clear this field to change the message to return "finished updating node pool"
		c.podToUpdateName = ""
	} else {
		c.podToUpdateName = pilotNameForStatefulSetReplica(ssCopy.Name, nextPartitionToUse)
	}

	_, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
	if err != nil {
		return err
	}

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
			return fmt.Errorf("pilot %q is version %s")
		}
	}
	return nil
}
