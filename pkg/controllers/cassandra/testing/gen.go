package testing

import (
	"fmt"
	"math/rand"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func FuzzCassandraNodePool(np *v1alpha1.CassandraClusterNodePool, rand *rand.Rand, size int) {
	np.Replicas = rand.Int31n(5)
}

func FuzzCassandraClusterNodePools(cluster *v1alpha1.CassandraCluster, rand *rand.Rand, size int) {
	if cluster.Spec.NodePools == nil {
		cluster.Spec.NodePools = []v1alpha1.CassandraClusterNodePool{}
	}
	if cluster.Status.NodePools == nil {
		cluster.Status.NodePools = map[string]v1alpha1.CassandraClusterNodePoolStatus{}
	}
	for i := 0; i < rand.Intn(5); i++ {
		np := v1alpha1.CassandraClusterNodePool{
			Name: fmt.Sprintf("np%d", i),
		}
		FuzzCassandraNodePool(&np, rand, size)
		nps := v1alpha1.CassandraClusterNodePoolStatus{
			ReadyReplicas: np.Replicas,
		}
		// 20% chance of ScaleOut
		if rand.Intn(4) == 0 {
			np.Replicas++
		}
		// 20% chance of ScaleIn
		if rand.Intn(4) == 0 {
			nps.ReadyReplicas++
		}
		// 20% chance of a NodePool removal
		if rand.Intn(4) != 0 {
			cluster.Spec.NodePools = append(cluster.Spec.NodePools, np)
		}
		// 20% chance of a NodePool create
		if rand.Intn(4) != 0 {
			cluster.Status.NodePools[np.Name] = nps
		}
	}
}
