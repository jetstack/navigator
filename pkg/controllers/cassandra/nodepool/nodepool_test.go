package nodepool_test

import (
	"testing"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestNodePoolControlSync(t *testing.T) {
	t.Run(
		"create a statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertStatefulSetsLength(1)
		},
	)
	t.Run(
		"ignore existing statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.AddObjectK(
				nodepool.StatefulSetForCluster(
					f.Cluster,
					&f.Cluster.Spec.NodePools[0],
				),
			)
			f.Run()
			f.AssertStatefulSetsLength(1)
		},
	)
}
