package rolebinding_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/rolebinding"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestRoleBindingSync(t *testing.T) {
	cluster1 := casstesting.ClusterForTest()
	rb1 := rolebinding.RoleBindingForCluster(cluster1)
	foreignRB1 := rb1.DeepCopy()
	foreignRB1.SetOwnerReferences([]v1.OwnerReference{})

	type testT struct {
		kubeObjects []runtime.Object
		cluster     *v1alpha1.CassandraCluster
		assertions  func(*testing.T, *controllers.State, testT)
		expectErr   bool
	}
	tests := map[string]testT{
		"create if not listed": {
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State, test testT) {
				expectedObject := rb1
				_, err := state.Clientset.
					RbacV1beta1().
					RoleBindings(expectedObject.Namespace).
					Get(expectedObject.Name, v1.GetOptions{})
				if err != nil {
					t.Error(err)
				}
			},
		},
		"no error if already listed": {
			kubeObjects: []runtime.Object{rb1},
			cluster:     cluster1,
		},
		"error if foreign owned": {
			kubeObjects: []runtime.Object{foreignRB1},
			cluster:     cluster1,
			expectErr:   true,
		},
	}

	for title, test := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				fixture := &framework.StateFixture{
					T:           t,
					KubeObjects: test.kubeObjects,
				}
				fixture.Start()
				defer fixture.Stop()
				state := fixture.State()
				c := rolebinding.NewControl(
					state.Clientset,
					state.RoleBindingLister,
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
