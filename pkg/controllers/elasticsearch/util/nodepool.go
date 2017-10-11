package util

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	hashutil "k8s.io/kubernetes/pkg/util/hash"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	NodePoolNameLabelKey      = "navigator.jetstack.io/elasticsearch-node-pool-name"
	NodePoolHashAnnotationKey = "navigator.jetstack.io/elasticsearch-node-pool-hash"
)

// ComputeHash returns a hash value calculated from pod template and a collisionCount to avoid hash collision
func ComputeNodePoolHash(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool, collisionCount *int32) string {
	hashVar := struct {
		Plugins    []string
		ESImage    v1alpha1.ElasticsearchImage
		PilotImage v1alpha1.ElasticsearchPilotImage
		Sysctl     []string
		NodePool   *v1alpha1.ElasticsearchClusterNodePool
	}{
		Plugins:    c.Spec.Plugins,
		ESImage:    c.Spec.Image,
		PilotImage: c.Spec.Pilot,
		Sysctl:     c.Spec.Sysctl,
		NodePool:   np,
	}

	hasher := fnv.New32a()
	hashutil.DeepHashObject(hasher, hashVar)

	// Add collisionCount in the hash if it exists.
	if collisionCount != nil {
		collisionCountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
		hasher.Write(collisionCountBytes)
	}

	return fmt.Sprintf("%s", hasher.Sum32())
}

func ClusterLabels(c *v1alpha1.ElasticsearchCluster) map[string]string {
	return map[string]string{
		"app":               "elasticsearch",
		ClusterNameLabelKey: c.Name,
	}
}

func NodePoolLabels(c *v1alpha1.ElasticsearchCluster, poolName string, roles ...v1alpha1.ElasticsearchClusterRole) map[string]string {
	labels := ClusterLabels(c)
	if poolName != "" {
		labels[NodePoolNameLabelKey] = poolName
	}
	for _, role := range roles {
		labels[string(role)] = "true"
	}
	return labels
}

func NodePoolResourceName(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) string {
	return fmt.Sprintf("%s-%s", ResourceBaseName(c), np.Name)
}

func SelectorForNodePool(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (labels.Selector, error) {
	nodePoolNameReq, err := labels.NewRequirement(NodePoolNameLabelKey, selection.Equals, []string{np.Name})
	if err != nil {
		return nil, err
	}
	clusterSelector, err := SelectorForCluster(c)
	if err != nil {
		return nil, err
	}
	return clusterSelector.Add(*nodePoolNameReq), nil
}

func PodControlledByCluster(c *v1alpha1.ElasticsearchCluster, pod *apiv1.Pod, ssLister appslisters.StatefulSetLister) (bool, error) {
	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef == nil || ownerRef.Kind != "StatefulSet" {
		return false, nil
	}
	ss, err := ssLister.StatefulSets(pod.Namespace).Get(ownerRef.Name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return metav1.IsControlledBy(ss, c), nil
}
