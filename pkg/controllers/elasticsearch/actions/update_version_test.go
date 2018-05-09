package actions

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
	apps "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/util/ptr"
)

const esImageRepo = "docker.elastic.co/elasticsearch/elasticsearch"

func TestUpdateVersion(t *testing.T) {
	type testT struct {
		kubeObjects         []runtime.Object
		navObjects          []runtime.Object
		cluster             *v1alpha1.ElasticsearchCluster
		nodePool            *v1alpha1.ElasticsearchClusterNodePool
		shouldUpdate        bool
		expectedStatefulSet *apps.StatefulSet
		err                 bool
	}
	redHealth := v1alpha1.ElasticsearchClusterHealthRed
	greenHealth := v1alpha1.ElasticsearchClusterHealthGreen
	tests := map[string]testT{
		"should update statefulset and set partition to update the highest ordinal pod": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{
					Name:            "es-test-data",
					Replicas:        ptr.Int32(3),
					Version:         "6.1.1",
					Image:           esImageRepo + ":6.1.1",
					CurrentRevision: "a",
					CurrentReplicas: 3,
					ReadyReplicas:   3,
				}),
			},
			navObjects: []runtime.Object{
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
			},
			cluster: generate.Cluster(generate.ClusterConfig{
				Name:    "test",
				Version: "6.1.2",
				Health:  &greenHealth,
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: true,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{
				Name:            "es-test-data",
				Replicas:        ptr.Int32(3),
				Version:         "6.1.1",
				Image:           esImageRepo + ":6.1.2",
				CurrentRevision: "a",
				CurrentReplicas: 3,
				Partition:       ptr.Int32(2),
				ReadyReplicas:   3,
			}),
			err: false,
		},
		"should update statefulset and set partition to update the second highest ordinal pod": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{
					Name:            "es-test-data",
					Replicas:        ptr.Int32(3),
					Version:         "6.1.1",
					Image:           esImageRepo + ":6.1.2",
					CurrentRevision: "a",
					CurrentReplicas: 2,
					UpdatedReplicas: 1,
					Partition:       ptr.Int32(2),
					ReadyReplicas:   3,
				}),
			},
			navObjects: []runtime.Object{
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Version: "6.1.2"}),
			},
			cluster: generate.Cluster(generate.ClusterConfig{
				Name:    "test",
				Version: "6.1.2",
				Health:  &greenHealth,
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: true,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{
				Name:            "es-test-data",
				Replicas:        ptr.Int32(3),
				Version:         "6.1.1",
				Image:           esImageRepo + ":6.1.2",
				CurrentRevision: "a",
				CurrentReplicas: 2,
				UpdatedReplicas: 1,
				Partition:       ptr.Int32(1),
				ReadyReplicas:   3,
			}),
			err: false,
		},
		"should not update a red cluster": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{
					Name:            "es-test-data",
					Replicas:        ptr.Int32(3),
					Version:         "6.1.1",
					Image:           esImageRepo + ":6.1.1",
					CurrentRevision: "a",
					ReadyReplicas:   3,
				}),
			},
			navObjects: []runtime.Object{
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
			},
			cluster: generate.Cluster(generate.ClusterConfig{
				Name:    "test",
				Version: "6.1.2",
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
				Health: &redHealth,
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: false,
			err:          false,
		},
		"should not update a node pool with an unready replica": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{
					Name:            "es-test-data",
					Replicas:        ptr.Int32(3),
					Version:         "6.1.1",
					Image:           esImageRepo + ":6.1.1",
					CurrentRevision: "a",
					ReadyReplicas:   2,
				}),
			},
			navObjects: []runtime.Object{
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
			},
			cluster: generate.Cluster(generate.ClusterConfig{
				Name:    "test",
				Version: "6.1.2",
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
				Health: &greenHealth,
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: false,
			err:          false,
		},
		"should set the updated version annotation on the statefulset when update is completed": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{
					Name:            "es-test-data",
					Replicas:        ptr.Int32(3),
					Version:         "6.1.1",
					Image:           esImageRepo + ":6.1.2",
					CurrentRevision: "b",
					UpdateRevision:  "b",
					CurrentReplicas: 3,
					Partition:       ptr.Int32(2),
					ReadyReplicas:   3,
				}),
			},
			navObjects: []runtime.Object{
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Version: "6.1.2"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Version: "6.1.2"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Version: "6.1.2"}),
			},
			cluster: generate.Cluster(generate.ClusterConfig{
				Name:    "test",
				Version: "6.1.2",
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
				Health: &greenHealth,
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: true,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{
				Name:            "es-test-data",
				Replicas:        ptr.Int32(3),
				Version:         "6.1.2",
				Image:           esImageRepo + ":6.1.2",
				CurrentRevision: "b",
				UpdateRevision:  "b",
				CurrentReplicas: 3,
				Partition:       ptr.Int32(2),
				ReadyReplicas:   3,
			}),
			err: false,
		},
		"should not update the next pod if the one before it hasn't finished updating": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{
					Name:            "es-test-data",
					Replicas:        ptr.Int32(3),
					Version:         "6.1.1",
					Image:           esImageRepo + ":6.1.2",
					CurrentRevision: "a",
					ReadyReplicas:   3,
					CurrentReplicas: 1,
					UpdatedReplicas: 2,
				}),
			},
			navObjects: []runtime.Object{
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Version: "6.1.1"}),
				generate.EsPilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Version: "6.1.2"}),
			},
			cluster: generate.Cluster(generate.ClusterConfig{
				Name:    "test",
				Version: "6.1.2",
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
				Health: &greenHealth,
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: false,
			err:          false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := &framework.StateFixture{
				T:                t,
				KubeObjects:      test.kubeObjects,
				NavigatorObjects: test.navObjects,
			}
			fixture.Start()
			defer fixture.Stop()

			state := fixture.State()
			updateVersion := &UpdateVersion{
				Cluster:  test.cluster,
				NodePool: test.nodePool,
			}
			err := updateVersion.Execute(state)
			if err != nil && !test.err {
				t.Errorf("Expected no error but got: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			actions := fixture.KubeClient().Actions()
			updateFound := false
			for _, action := range actions {
				if action.Matches("update", "statefulsets") {
					updateFound = true
				}
			}
			if !test.shouldUpdate && updateFound {
				t.Errorf("Update to statefulset performed when it should not have")
			}
			if test.shouldUpdate && !updateFound {
				t.Errorf("Update to statefulset not performed when it should have been")
			}
			if test.expectedStatefulSet != nil {
				actualStatefulSet, err := fixture.KubeClient().AppsV1beta1().StatefulSets(test.expectedStatefulSet.Namespace).Get(test.expectedStatefulSet.Name, metav1.GetOptions{})
				if err != nil {
					t.Errorf("Got error when retrieving statefulset: %v", err)
					t.Fail()
				}
				if !reflect.DeepEqual(test.expectedStatefulSet, actualStatefulSet) {
					t.Errorf("Expected did not equal actual: %s", pretty.Diff(actualStatefulSet, test.expectedStatefulSet))
				}
			}
		})
	}
}
