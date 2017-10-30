package util

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	hashutil "github.com/jetstack/navigator/pkg/util/hash"
)

const (
	typeName                  = "cass"
	kindName                  = "CassandraCluster"
	ClusterNameLabelKey       = "navigator.jetstack.io/cassandra-cluster-name"
	NodePoolNameLabelKey      = "navigator.jetstack.io/cassandra-node-pool-name"
	NodePoolHashAnnotationKey = "navigator.jetstack.io/cassandra-node-pool-hash"
)

func NewControllerRef(c *v1alpha1.CassandraCluster) metav1.OwnerReference {
	return *metav1.NewControllerRef(c, schema.GroupVersionKind{
		Group:   navigator.GroupName,
		Version: "v1alpha1",
		Kind:    kindName,
	})
}

func ResourceBaseName(c *v1alpha1.CassandraCluster) string {
	return typeName + "-" + c.Name
}

func NodePoolResourceName(c *v1alpha1.CassandraCluster, np *v1alpha1.CassandraClusterNodePool) string {
	return fmt.Sprintf("%s-%s", ResourceBaseName(c), np.Name)
}

func ClusterLabels(c *v1alpha1.CassandraCluster) map[string]string {
	return map[string]string{
		"app":               "cassandracluster",
		ClusterNameLabelKey: c.Name,
	}
}

func SelectorForCluster(c *v1alpha1.CassandraCluster) (labels.Selector, error) {
	clusterNameReq, err := labels.NewRequirement(
		ClusterNameLabelKey,
		selection.Equals,
		[]string{c.Name},
	)
	if err != nil {
		return nil, err
	}
	return labels.NewSelector().Add(*clusterNameReq), nil
}

func NodePoolLabels(
	c *v1alpha1.CassandraCluster,
	poolName string,
) map[string]string {
	labels := ClusterLabels(c)
	if poolName != "" {
		labels[NodePoolNameLabelKey] = poolName
	}
	return labels
}

func Int32Ptr(i int32) *int32 {
	return &i
}

// ComputeHash returns a hash value calculated from pod template and a
// collisionCount to avoid hash collision
func ComputeNodePoolHash(
	c *v1alpha1.CassandraCluster,
	np *v1alpha1.CassandraClusterNodePool,
	collisionCount *int32,
) string {
	hashVar := struct {
		// Plugins    []string
		// ESImage    v1alpha1.ElasticsearchImage
		// PilotImage v1alpha1.ElasticsearchPilotImage
		Sysctl   []string
		NodePool *v1alpha1.CassandraClusterNodePool
	}{
		// Plugins:    c.Spec.Plugins,
		// ESImage:    c.Spec.Image,
		// PilotImage: c.Spec.Pilot,
		Sysctl:   c.Spec.Sysctl,
		NodePool: np,
	}

	hasher := fnv.New32a()
	hashutil.DeepHashObject(hasher, hashVar)

	// Add collisionCount in the hash if it exists.
	if collisionCount != nil {
		collisionCountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
		hasher.Write(collisionCountBytes)
	}

	return fmt.Sprintf("%d", hasher.Sum32())
}
