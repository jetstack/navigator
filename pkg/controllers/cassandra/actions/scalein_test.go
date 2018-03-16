package actions_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
)

func TestScaleIn(t *testing.T) {
	type testT struct {
		kubeObjects         []runtime.Object
		navObjects          []runtime.Object
		cluster             generate.CassandraClusterConfig
		nodePool            generate.CassandraClusterNodePoolConfig
		expectedStatefulSet *generate.StatefulSetConfig
		expectedErr         bool
		mutator             func(*framework.StateFixture)
	}
	tests := map[string]testT{
		"Error if StatefulSet not listed": {
			cluster: generate.CassandraClusterConfig{
				Name:      "cluster1",
				Namespace: "ns1",
			},
			nodePool: generate.CassandraClusterNodePoolConfig{
				Name:     "pool1",
				Replicas: 123,
			},
			expectedErr: true,
		},
		"Error if clientset.Update fails (e.g. listed but not found)": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(
					generate.StatefulSetConfig{
						Name:      "cass-cluster1-pool1",
						Namespace: "ns1",
						Replicas:  int32Ptr(122),
					},
				),
			},
			cluster: generate.CassandraClusterConfig{
				Name:      "cluster1",
				Namespace: "ns1",
			},
			nodePool: generate.CassandraClusterNodePoolConfig{
				Name:     "pool1",
				Replicas: 123,
			},
			expectedErr: true,
			mutator: func(f *framework.StateFixture) {
				err := f.KubeClient().
					AppsV1beta1().
					StatefulSets("ns1").
					Delete("cass-cluster1-pool1", &metav1.DeleteOptions{})
				if err != nil {
					f.T.Fatal(err)
				}
			},
		},
		"Idempotent: No error if ReplicaCount already matches the actual ReplicaCount": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(
					generate.StatefulSetConfig{
						Name:      "cass-cluster1-pool1",
						Namespace: "ns1",
						Replicas:  int32Ptr(124),
					},
				),
			},
			cluster: generate.CassandraClusterConfig{
				Name:      "cluster1",
				Namespace: "ns1",
			},
			nodePool: generate.CassandraClusterNodePoolConfig{
				Name:     "pool1",
				Replicas: 124,
			},
			expectedStatefulSet: &generate.StatefulSetConfig{
				Name:      "cass-cluster1-pool1",
				Namespace: "ns1",
				Replicas:  int32Ptr(124),
			},
			expectedErr: false,
		},
		"The replicas count is decremented": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(
					generate.StatefulSetConfig{
						Name:      "cass-cluster1-pool1",
						Namespace: "ns1",
						Replicas:  int32Ptr(125),
					},
				),
			},
			cluster: generate.CassandraClusterConfig{
				Name:      "cluster1",
				Namespace: "ns1",
			},
			nodePool: generate.CassandraClusterNodePoolConfig{
				Name:     "pool1",
				Replicas: 120,
			},
			expectedStatefulSet: &generate.StatefulSetConfig{
				Name:      "cass-cluster1-pool1",
				Namespace: "ns1",
				Replicas:  int32Ptr(120),
			},
		},
	}

	for name, test := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				fixture := &framework.StateFixture{
					T:                t,
					KubeObjects:      test.kubeObjects,
					NavigatorObjects: test.navObjects,
				}
				fixture.Start()
				defer fixture.Stop()
				state := fixture.State()
				if test.mutator != nil {
					test.mutator(fixture)
				}
				a := &actions.ScaleIn{
					Cluster:  generate.CassandraCluster(test.cluster),
					NodePool: generate.CassandraClusterNodePool(test.nodePool),
				}
				err := a.Execute(state)
				if err != nil {
					t.Logf("The error returned by Execute was: %s", err)
				}
				if !test.expectedErr && err != nil {
					t.Errorf("Unexpected error: %s", err)
				}
				if test.expectedErr && err == nil {
					t.Errorf("Expected an error")
				}
				if test.expectedStatefulSet != nil {
					actualStatefulSet, err := fixture.KubeClient().
						AppsV1beta1().
						StatefulSets(test.expectedStatefulSet.Namespace).
						Get(test.expectedStatefulSet.Name, metav1.GetOptions{})
					if err != nil {
						t.Fatalf("Unexpected error retrieving statefulset: %v", err)
					}
					generate.AssertStatefulSetMatches(t, *test.expectedStatefulSet, actualStatefulSet)
				}
			},
		)
	}
}
