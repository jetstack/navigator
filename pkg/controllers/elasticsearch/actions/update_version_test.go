package actions

import (
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/kr/pretty"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

const esImageRepo = "docker.elastic.co/elasticsearch/elasticsearch"

func int32Ptr(i int32) *int32 {
	return &i
}

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
	tests := map[string]testT{
		"should update statefulset and set partition to update the highest ordinal pod": {
			kubeObjects: []runtime.Object{
				generateStatefulSet(statefulSetGeneratorConfig{
					name:            "es-test-data",
					replicas:        int32Ptr(3),
					version:         "6.1.1",
					image:           esImageRepo + ":6.1.1",
					currentRevision: "a",
					currentReplicas: 3,
				}),
			},
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 1),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 2),
			},
			cluster:      clusterWithVersionNodePools("test", "6.1.2", nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData)),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: true,
			expectedStatefulSet: generateStatefulSet(statefulSetGeneratorConfig{
				name:            "es-test-data",
				replicas:        int32Ptr(3),
				version:         "6.1.1",
				image:           esImageRepo + ":6.1.2",
				currentRevision: "a",
				currentReplicas: 3,
				partition:       int32Ptr(2),
			}),
			err: false,
		},
		"should update statefulset and set partition to update the second highest ordinal pod": {
			kubeObjects: []runtime.Object{
				generateStatefulSet(statefulSetGeneratorConfig{
					name:            "es-test-data",
					replicas:        int32Ptr(3),
					version:         "6.1.1",
					image:           esImageRepo + ":6.1.1",
					currentRevision: "a",
					currentReplicas: 2,
					updatedReplicas: 1,
					partition:       int32Ptr(2),
				}),
			},
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 1),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 2),
			},
			cluster: generateCluster(clusterGeneratorConfig{
				name:    "test",
				version: "6.1.2",
				nodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: true,
			expectedStatefulSet: generateStatefulSet(statefulSetGeneratorConfig{
				name:            "es-test-data",
				replicas:        int32Ptr(3),
				version:         "6.1.1",
				image:           esImageRepo + ":6.1.2",
				currentRevision: "a",
				currentReplicas: 2,
				updatedReplicas: 1,
				partition:       int32Ptr(1),
			}),
			err: false,
		},
		"should not update a red cluster": {
			kubeObjects: []runtime.Object{
				generateStatefulSet(statefulSetGeneratorConfig{
					name:            "es-test-data",
					replicas:        int32Ptr(3),
					version:         "6.1.1",
					image:           esImageRepo + ":6.1.1",
					currentRevision: "a",
				}),
			},
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 1),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 2),
			},
			cluster: generateCluster(clusterGeneratorConfig{
				name:    "test",
				version: "6.1.2",
				nodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
				health: v1alpha1.ElasticsearchClusterHealthRed,
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: false,
			expectedStatefulSet: generateStatefulSet(statefulSetGeneratorConfig{
				name:            "es-test-data",
				replicas:        int32Ptr(3),
				version:         "6.1.1",
				image:           esImageRepo + ":6.1.1",
				currentRevision: "a",
			}),
			err: true,
		},
		"should set the updated version annotation on the statefulset when update is completed": {
			kubeObjects: []runtime.Object{
				generateStatefulSet(statefulSetGeneratorConfig{
					name:            "es-test-data",
					replicas:        int32Ptr(3),
					version:         "6.1.1",
					image:           esImageRepo + ":6.1.2",
					currentRevision: "b",
					updateRevision:  "b",
					currentReplicas: 3,
					partition:       int32Ptr(2),
				}),
			},
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 1),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 2),
			},
			cluster: generateCluster(clusterGeneratorConfig{
				name:    "test",
				version: "6.1.2",
				nodePools: []v1alpha1.ElasticsearchClusterNodePool{
					nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
				},
			}),
			nodePool:     nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			shouldUpdate: true,
			expectedStatefulSet: generateStatefulSet(statefulSetGeneratorConfig{
				name:            "es-test-data",
				replicas:        int32Ptr(3),
				version:         "6.1.2",
				image:           esImageRepo + ":6.1.2",
				currentRevision: "b",
				updateRevision:  "b",
				currentReplicas: 3,
				partition:       int32Ptr(2),
			}),
			err: false,
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
			state := fixture.State()
			scale := &UpdateVersion{
				Cluster:  test.cluster,
				NodePool: test.nodePool,
			}
			err := scale.Execute(state)
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

type clusterGeneratorConfig struct {
	name          string
	nodePools     []v1alpha1.ElasticsearchClusterNodePool
	version       string
	clusterConfig v1alpha1.NavigatorClusterConfig
	health        v1alpha1.ElasticsearchClusterHealth
}

func generateCluster(c clusterGeneratorConfig) *v1alpha1.ElasticsearchCluster {
	return &v1alpha1.ElasticsearchCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: "default",
		},
		Spec: v1alpha1.ElasticsearchClusterSpec{
			NavigatorClusterConfig: c.clusterConfig,
			Version:                *semver.New(c.version),
			NodePools:              c.nodePools,
		},
		Status: v1alpha1.ElasticsearchClusterStatus{
			Health: c.health,
		},
	}
}

type statefulSetGeneratorConfig struct {
	name                             string
	replicas                         *int32
	version, image                   string
	partition                        *int32
	currentRevision, updateRevision  string
	currentReplicas, updatedReplicas int32
}

func generateStatefulSet(c statefulSetGeneratorConfig) *apps.StatefulSet {
	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: "default",
			Annotations: map[string]string{
				v1alpha1.ElasticsearchNodePoolVersionAnnotation: c.version,
			},
		},
		Spec: apps.StatefulSetSpec{
			Replicas: c.replicas,
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				RollingUpdate: &apps.RollingUpdateStatefulSetStrategy{
					Partition: c.partition,
				},
			},
			Template: core.PodTemplateSpec{
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Image: c.image,
						},
					},
				},
			},
		},
		Status: apps.StatefulSetStatus{
			CurrentRevision: c.currentRevision,
			UpdateRevision:  c.updateRevision,
			CurrentReplicas: c.currentReplicas,
			UpdatedReplicas: c.updatedReplicas,
		},
	}
}
