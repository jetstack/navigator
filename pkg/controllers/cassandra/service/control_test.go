package service_test

import (
	"testing"

	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
)

func TestSync(t *testing.T) {
	cluster1 := casstesting.ClusterForTest()
	service1 := service.NodesServiceForCluster(cluster1)
	serviceFactory := service.NodesServiceForCluster
	foreignService1 := service1.DeepCopy()
	foreignService1.SetOwnerReferences([]v1.OwnerReference{})

	type testT struct {
		kubeObjects []runtime.Object
		cluster     *v1alpha1.CassandraCluster
		assertions  func(*testing.T, *controllers.State, testT)
		expectErr   bool
	}

	tests := map[string]testT{
		"create service": {
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State, test testT) {
				expectedService := service1
				_, err := state.Clientset.
					CoreV1().
					Services(expectedService.Namespace).
					Get(expectedService.Name, v1.GetOptions{})
				if err != nil {
					t.Error(err)
				}
			},
		},

		"service exists": {
			kubeObjects: []runtime.Object{service1},
			cluster:     cluster1,
		},
		"error if foreign owned": {
			kubeObjects: []runtime.Object{foreignService1},
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
				c := service.NewControl(
					state.Clientset,
					state.ServiceLister,
					state.Recorder,
					serviceFactory,
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
