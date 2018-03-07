package pilot_test

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"

	"github.com/kr/pretty"

	"github.com/jetstack/navigator/pkg/cassandra/version"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

func clusterPod(cluster *v1alpha1.CassandraCluster, name string) *v1.Pod {
	pod := &v1.Pod{}
	pod.SetName(name)
	pod.SetNamespace(cluster.GetNamespace())
	pod.SetOwnerReferences(
		[]metav1.OwnerReference{
			util.NewControllerRef(cluster),
		},
	)
	return pod
}

func TestPilotSync(t *testing.T) {
	cluster1 := casstesting.ClusterForTest()
	cluster1pod1 := clusterPod(cluster1, "c1p1")
	cluster1pod2 := clusterPod(cluster1, "c1p2")
	cluster1pilot1 := pilot.PilotForCluster(cluster1, cluster1pod1)
	cluster1pilot1foreign := cluster1pilot1.DeepCopy()
	cluster1pilot1foreign.SetOwnerReferences([]metav1.OwnerReference{})

	cluster2 := casstesting.ClusterForTest()
	cluster2.SetName("cluster2")
	cluster2.SetUID("uid2")
	cluster2pod1 := clusterPod(cluster2, "c2p1")

	type testT struct {
		kubeObjects []runtime.Object
		navObjects  []runtime.Object
		cluster     *v1alpha1.CassandraCluster
		assertions  func(*testing.T, *controllers.State)
		expectErr   bool
	}

	tests := map[string]testT{
		"each cluster pod gets a pilot": {
			kubeObjects: []runtime.Object{
				cluster1pod1,
				cluster1pod2,
				cluster2pod1,
			},
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State) {
				pilots, err := state.NavigatorClientset.
					Navigator().Pilots(cluster1.Namespace).List(metav1.ListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expectedPilotCount := 2
				pilotCount := len(pilots.Items)
				if pilotCount != expectedPilotCount {
					t.Log(pilots.Items)
					t.Errorf("Unexpected pilot count: %d != %d", expectedPilotCount, pilotCount)
				}
			},
		},
		"non-cluster pods are ignored": {
			kubeObjects: []runtime.Object{
				cluster1pod1,
				cluster2pod1,
			},
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State) {
				pilots, err := state.NavigatorClientset.
					Navigator().Pilots(cluster1.Namespace).List(metav1.ListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expectedPilotCount := 1
				pilotCount := len(pilots.Items)
				if pilotCount != expectedPilotCount {
					t.Log(pilots.Items)
					t.Errorf("Unexpected pilot count: %d != %d", expectedPilotCount, pilotCount)
				}
			},
		},
		"no error if pilot exists": {
			kubeObjects: []runtime.Object{cluster1pod1},
			navObjects:  []runtime.Object{cluster1pilot1},
			cluster:     cluster1,
		},
		"error if foreign owned": {
			kubeObjects: []runtime.Object{cluster1pod1},
			navObjects:  []runtime.Object{cluster1pilot1foreign},
			cluster:     cluster1,
			expectErr:   true,
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
				state := fixture.State()
				c := pilot.NewControl(
					state.NavigatorClientset,
					state.PilotLister,
					state.PodLister,
					state.StatefulSetLister,
					state.Recorder,
				)
				err := c.Sync(test.cluster)
				if err != nil {
					if !test.expectErr {
						t.Errorf("Unexpected error: %s", err)
					}
				} else {
					if test.expectErr {
						t.Error("Missing error")
					}
				}
				if test.assertions != nil {
					test.assertions(t, state)
				}
			},
		)
	}
}

func AssertClusterEqual(t *testing.T, c1, c2 *v1alpha1.CassandraCluster) {
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("Clusters are not equal: %s", pretty.Diff(c1, c2))
	}
}

func TestStatusUpdate(t *testing.T) {
	type testT struct {
		kubeObjects []runtime.Object
		navObjects  []runtime.Object
		cluster     *v1alpha1.CassandraCluster
		assertions  func(t *testing.T, original, updated *v1alpha1.CassandraCluster)
		expectErr   bool
	}
	cluster := casstesting.ClusterForTest()
	tests := map[string]testT{
		"no matching pilots": {
			navObjects: []runtime.Object{
				pilot.NewPilotBuilder().Build(),
			},
			cluster:    cluster,
			assertions: AssertClusterEqual,
		},
		"nil cassandra status": {
			navObjects: []runtime.Object{
				pilot.NewPilotBuilder().ForCluster(cluster).Build(),
			},
			cluster:    cluster,
			assertions: AssertClusterEqual,
		},
		"nil cassandra version": {
			navObjects: []runtime.Object{
				pilot.NewPilotBuilder().
					ForCluster(cluster).
					WithCassandraStatus().
					Build(),
			},
			cluster:    cluster,
			assertions: AssertClusterEqual,
		},
		"missing nodepool label": {
			navObjects: []runtime.Object{
				pilot.NewPilotBuilder().
					ForCluster(cluster).
					WithDiscoveredCassandraVersion("3.11.2").
					Build(),
			},
			cluster:    cluster,
			assertions: AssertClusterEqual,
		},
		"set version if missing": {
			navObjects: []runtime.Object{
				pilot.NewPilotBuilder().
					ForCluster(cluster).
					ForNodePool(&cluster.Spec.NodePools[0]).
					WithDiscoveredCassandraVersion("3.11.2").
					Build(),
			},
			cluster: cluster,
			assertions: func(t *testing.T, inCluster, outCluster *v1alpha1.CassandraCluster) {
				expectedVersion := version.New("3.11.2")
				actualVersion := outCluster.Status.NodePools["RingNodes"].Version
				if !expectedVersion.Equal(actualVersion) {
					t.Errorf("Version mismatch. Expected %s != %s", expectedVersion, actualVersion)
				}
			},
		},
		"set version if lower": {
			navObjects: []runtime.Object{
				pilot.NewPilotBuilder().
					ForCluster(cluster).
					ForNodePool(&cluster.Spec.NodePools[0]).
					WithDiscoveredCassandraVersion("3.11.2").
					Build(),
			},
			cluster: cluster,
			assertions: func(t *testing.T, inCluster, outCluster *v1alpha1.CassandraCluster) {
				expectedVersion := version.New("3.11.2")
				actualVersion := outCluster.Status.NodePools["RingNodes"].Version
				if !expectedVersion.Equal(actualVersion) {
					t.Errorf("Version mismatch. Expected %s != %s", expectedVersion, actualVersion)
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
				state := fixture.State()
				c := pilot.NewControl(
					state.NavigatorClientset,
					state.PilotLister,
					state.PodLister,
					state.StatefulSetLister,
					state.Recorder,
				)
				cluster = test.cluster.DeepCopy()
				err := c.Sync(cluster)
				if err != nil {
					if !test.expectErr {
						t.Errorf("Unexpected error: %s", err)
					}
				} else {
					if test.expectErr {
						t.Error("Missing error")
					}
				}
				if test.assertions != nil {
					test.assertions(t, test.cluster, cluster)
				}
			},
		)
	}
}
