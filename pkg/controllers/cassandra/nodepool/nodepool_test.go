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
	t.Run(
		"update statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			unsyncedSet := nodepool.StatefulSetForCluster(
				f.Cluster,
				&f.Cluster.Spec.NodePools[0],
			)
			unsyncedSet.SetLabels(map[string]string{})
			f.AddObjectK(unsyncedSet)
			f.Run()
			f.AssertStatefulSetsLength(1)
			sets := f.StatefulSets()
			set := sets.Items[0]
			labels := set.GetLabels()
			if len(labels) == 0 {
				t.Log(set)
				t.Error("StatefulSet was not updated")
			}
		},
	)
	t.Run(
		"error on update foreign statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			foreignUnsyncedSet := nodepool.StatefulSetForCluster(
				f.Cluster,
				&f.Cluster.Spec.NodePools[0],
			)
			foreignUnsyncedSet.SetLabels(map[string]string{})
			foreignUnsyncedSet.OwnerReferences = nil
			f.AddObjectK(foreignUnsyncedSet)
			f.RunExpectError()
		},
	)
}
