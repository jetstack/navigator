package actions

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/golang/glog"
	apps "k8s.io/api/apps/v1beta1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

type UpdateVersion struct {
	// The Elasticsearch cluster resource being operated on
	Cluster *v1alpha1.ElasticsearchCluster
	// The node pool being scaled
	NodePool *v1alpha1.ElasticsearchClusterNodePool
}

var _ controllers.Action = &UpdateVersion{}

func (c *UpdateVersion) Name() string {
	return "UpdateVersion"
}

func (m *UpdateVersion) Execute(state *controllers.State) error {
	if m.NodePool == nil || m.Cluster == nil {
		return fmt.Errorf("internal error: nodepool and cluster must be set")
	}

	statefulSetName := util.NodePoolResourceName(m.Cluster, m.NodePool)
	statefulSet, err := state.StatefulSetLister.StatefulSets(m.Cluster.Namespace).Get(statefulSetName)
	if err != nil {
		return err
	}

	// TODO: ensure shard reallocation is disabled
	if m.Cluster.Status.Health == v1alpha1.ElasticsearchClusterHealthRed {
		return fmt.Errorf("cluster is in a red state, refusing to upgrade node pool %q", m.NodePool.Name)
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
	if m.Cluster.Spec.Version.Equal(*currentVersion) {
		return nil
	}

	newImageSpec, err := nodepool.ESImageToUse(&m.Cluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot determine new image details to use: %v", err)
	}

	ssCopy := statefulSet.DeepCopy()
	// check if the pod templates image field needs updating
	newImage := newImageSpec.Repository + ":" + newImageSpec.Tag
	if len(ssCopy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("internal error: no containers specified on statefulset pod template")
	}
	existingImage := ssCopy.Spec.Template.Spec.Containers[0].Image
	if existingImage != newImage {
		if ssCopy.Spec.UpdateStrategy.RollingUpdate == nil {
			ssCopy.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateStatefulSetStrategy{}
		}
		// we update the image, but don't allow any pods to be updated yet
		ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition = ssCopy.Spec.Replicas
		ssCopy.Spec.Template.Spec.Containers[0].Image = newImage
		ssCopy, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
		if err != nil {
			return err
		}
	}

	desiredReplicasPtr := ssCopy.Spec.Replicas
	if desiredReplicasPtr == nil {
		return fmt.Errorf("desired replicas on statefulset cannot be nil")
	}
	currentPartition := int32(0)
	if ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		currentPartition = *ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition
	}
	desiredReplicas := *desiredReplicasPtr
	updatedReplicas := ssCopy.Status.UpdatedReplicas
	currentReplicas := ssCopy.Status.CurrentReplicas
	glog.V(4).Infof("ElasticsearchCluster node pool %q/%q - desired replicas: %d, updated replicas: %d, current replicas: %d, current partition: %d", m.Cluster.Name, m.NodePool.Name, desiredReplicas, updatedReplicas, currentReplicas, currentPartition)

	// we subtract 1 here because we count from 0 when counting replicas
	nextPartitionToUse := int32(desiredReplicas - updatedReplicas - 1)

	// we have already ensured the cluster is not in a red state near the start of the function
	// TODO: check that the pilot previously being updated is now healthy again?

	// we are ready to reduce the partition number and begin updating the next
	// replica
	if currentPartition != nextPartitionToUse && nextPartitionToUse >= 0 {
		if ssCopy.Spec.UpdateStrategy.RollingUpdate == nil {
			ssCopy.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateStatefulSetStrategy{}
		}
		ssCopy.Spec.UpdateStrategy.RollingUpdate.Partition = &nextPartitionToUse
		ssCopy, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
		if err != nil {
			return err
		}
	}

	if ssCopy.Status.CurrentRevision == ssCopy.Status.UpdateRevision {
		// we only set this once the upgrade is complete for the node pool
		ssCopy.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation] = m.Cluster.Spec.Version.String()
		ssCopy, err = state.Clientset.AppsV1beta1().StatefulSets(ssCopy.Namespace).Update(ssCopy)
		if err != nil {
			return err
		}
	}

	return nil
}
