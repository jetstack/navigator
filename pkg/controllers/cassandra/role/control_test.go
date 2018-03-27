package role_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/role"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestRoleSync(t *testing.T) {
	cluster1 := casstesting.ClusterForTest()
	role1 := role.RoleForCluster(cluster1)
	foreignRole1 := role1.DeepCopy()
	foreignRole1.SetOwnerReferences([]v1.OwnerReference{})

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
				expectedObject := role1
				_, err := state.Clientset.
					RbacV1beta1().
					Roles(expectedObject.Namespace).
					Get(expectedObject.Name, v1.GetOptions{})
				if err != nil {
					t.Error(err)
				}
			},
		},
		"no error if already listed": {
			kubeObjects: []runtime.Object{role1},
			cluster:     cluster1,
		},
		"error if foreign owned": {
			kubeObjects: []runtime.Object{foreignRole1},
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
				c := role.NewControl(
					state.Clientset,
					state.RoleLister,
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
