package testing

import (
	"fmt"
	"math/rand"

	"github.com/jetstack/navigator/pkg/api/version"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/util/ptr"
)

var versions = []*version.Version{
	version.New("2.0.0"),
	version.New("3.11"),
	version.New("3.11.1"),
	version.New("3.11.2"),
	version.New("4.0.0"),
}

func FuzzCassandraCluster(cluster *v1alpha1.CassandraCluster, rand *rand.Rand, size int) {
	cluster.Spec.Version = *versions[rand.Intn(len(versions))]
	FuzzCassandraClusterNodePools(cluster, rand, size)
	// 20% chance of patch upgrade
	if rand.Intn(4) == 0 {
		cluster.Spec.Version = *cluster.Spec.Version.BumpPatch()
	}
	// 20% chance of minor upgrade
	if rand.Intn(4) == 0 {
		cluster.Spec.Version = *cluster.Spec.Version.BumpMinor()
	}
	// 20% chance of major upgrade
	if rand.Intn(4) == 0 {
		cluster.Spec.Version = *cluster.Spec.Version.BumpMajor()
	}
}

func FuzzCassandraNodePool(np *v1alpha1.CassandraClusterNodePool, rand *rand.Rand, size int) {
	np.Replicas = ptr.Int32(rand.Int31n(5))
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
			ReadyReplicas: *np.Replicas,
			Version:       version.New(cluster.Spec.Version.String()),
		}
		// 20% chance of too new version
		if rand.Intn(4) == 0 {
			nps.Version = nps.Version.BumpMajor()
		}
		// 20% chance of unreported version
		if rand.Intn(4) == 0 {
			nps.Version = nil
		}
		// 20% chance of ScaleOut
		if rand.Intn(4) == 0 {
			np.Replicas = ptr.Int32(*np.Replicas + 1)
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
