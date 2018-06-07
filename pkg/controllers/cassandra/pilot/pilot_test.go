package pilot_test

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/api/version"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

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
				actualVersion := outCluster.Status.NodePools["region-1-zone-a"].Version
				if actualVersion == nil || !expectedVersion.Equal(actualVersion) {
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
				actualVersion := outCluster.Status.NodePools["region-1-zone-a"].Version
				if actualVersion == nil || !expectedVersion.Equal(actualVersion) {
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
