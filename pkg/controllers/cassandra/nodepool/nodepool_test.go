package nodepool_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestNodePoolControlSync(t *testing.T) {
	cluster1 := casstesting.ClusterForTest()
	set1 := nodepool.StatefulSetForCluster(cluster1, &cluster1.Spec.NodePools[0])

	type testT struct {
		kubeObjects        []runtime.Object
		navObjects         []runtime.Object
		cluster            *v1alpha1.CassandraCluster
		fixtureManipulator func(*testing.T, *framework.StateFixture)
		assertions         func(*testing.T, *controllers.State, testT)
		expectErr          bool
	}

	tests := map[string]testT{
		"create if not exists": {
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State, test testT) {
				expectedObject := set1
				_, err := state.Clientset.AppsV1beta1().
					StatefulSets(expectedObject.Namespace).
					Get(expectedObject.Name, v1.GetOptions{})
				if err != nil {
					t.Error(err)
				}
			},
		},

		"service exists": {
			kubeObjects: []runtime.Object{set1},
			cluster:     cluster1,
		},
		"not yet listed": {
			kubeObjects: []runtime.Object{},
			cluster:     cluster1,
			fixtureManipulator: func(t *testing.T, fixture *framework.StateFixture) {
				_, err := fixture.KubeClient().AppsV1beta1().StatefulSets(set1.Namespace).Create(set1)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for title, test := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				fixture := &framework.StateFixture{
					T:                t,
					KubeObjects:      test.kubeObjects,
					NavigatorObjects: test.navObjects,
				}
				fixture.Start()
				defer fixture.Stop()
				if test.fixtureManipulator != nil {
					test.fixtureManipulator(t, fixture)
				}
				state := fixture.State()
				c := nodepool.NewControl(
					state.Clientset,
					state.StatefulSetLister,
					state.Recorder,
				)
				err := c.Sync(test.cluster)
				if err == nil {
					if test.expectErr {
						t.Error("Expected an error")
					}
				} else {
					if !test.expectErr {
						t.Error(err)
					}
				}
				if test.assertions != nil {
					test.assertions(t, state, test)
				}
			},
		)
	}
}
