package actions

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/kr/pretty"
	apps "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	testutil "github.com/jetstack/navigator/internal/test/util"
	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func TestScale(t *testing.T) {
	type testT struct {
		kubeObjects         []runtime.Object
		navObjects          []runtime.Object
		cluster             *v1alpha1.ElasticsearchCluster
		nodePool            *v1alpha1.ElasticsearchClusterNodePool
		replicas            int32
		shouldUpdate        bool
		expectedStatefulSet *apps.StatefulSet
		err                 bool
	}
	tests := map[string]testT{
		"should not scale statefulset if documents still remain": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			},
			navObjects: []runtime.Object{
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Documents: int64Ptr(2)}),
			},
			cluster:             clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:            nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			replicas:            2,
			shouldUpdate:        false,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			err:                 false,
		},
		"should scale statefulset if no documents remain": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			},
			navObjects: []runtime.Object{
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Documents: int64Ptr(1)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Documents: int64Ptr(1)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
			},
			cluster:             clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:            nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			replicas:            2,
			shouldUpdate:        true,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(2)}),
			err:                 false,
		},
		"should not scale statefulset if a pilot that should exist is missing": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			},
			navObjects: []runtime.Object{
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
			},
			cluster:             clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:            nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			replicas:            2,
			shouldUpdate:        false,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			err:                 false,
		},
		"should not error if statefulset doesn't exist": {
			kubeObjects: []runtime.Object{},
			navObjects: []runtime.Object{
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
			},
			cluster:             clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:            nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			replicas:            2,
			shouldUpdate:        false,
			expectedStatefulSet: nil,
			err:                 false,
		},
		"should not update if replica difference is zero": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			},
			navObjects: []runtime.Object{
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-0", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-1", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
				generate.Pilot(generate.PilotConfig{Name: "es-test-data-2", Cluster: "test", NodePool: "data", Documents: int64Ptr(0)}),
			},
			cluster:             clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData)),
			nodePool:            nodePoolPtrWithNameReplicasRoles("data", 3, v1alpha1.ElasticsearchRoleData),
			replicas:            3,
			shouldUpdate:        false,
			expectedStatefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			err:                 false,
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
			scale := &Scale{
				Cluster:  test.cluster,
				NodePool: test.nodePool,
				Replicas: test.replicas,
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
func TestPilotsForStatefulSet(t *testing.T) {
	type testT struct {
		navObjects     []runtime.Object
		cluster        *v1alpha1.ElasticsearchCluster
		nodePool       *v1alpha1.ElasticsearchClusterNodePool
		statefulSet    *apps.StatefulSet
		expectedPilots []*v1alpha1.Pilot
		err            bool
	}
	tests := []testT{
		{
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 0),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			expectedPilots: []*v1alpha1.Pilot{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 0),
			},
		},
		{
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 10),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			err:         true,
		},
		{
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-master-0", "test", "master", 10),
				pilotWithNameDocuments("es-test-master-1", "test", "master", 10),
				pilotWithNameDocuments("es-test-master-2", "test", "master", 0),
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 1),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 2),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData), nodePoolWithNameReplicasRoles("master", 3, v1alpha1.ElasticsearchRoleMaster)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			expectedPilots: []*v1alpha1.Pilot{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 1),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 2),
			},
			err: false,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			fixture := &framework.StateFixture{
				T:                t,
				NavigatorObjects: test.navObjects,
			}
			fixture.Start()
			state := fixture.State()
			pilots, err := pilotsForStatefulSet(state, test.cluster, test.nodePool, test.statefulSet)
			if err != nil && !test.err {
				t.Errorf("Expected no error but got: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			if len(pilots) != len(test.expectedPilots) ||
				!testutil.ContainsAll(pilots, test.expectedPilots) ||
				!testutil.ContainsAll(test.expectedPilots, pilots) {
				t.Errorf("Expected did not equal actual: %s", pretty.Diff(pilots, test.expectedPilots))
			}
		})
	}
}

func TestCanScaleNodePool(t *testing.T) {
	type testT struct {
		kubeObjects []runtime.Object
		navObjects  []runtime.Object
		cluster     *v1alpha1.ElasticsearchCluster
		nodePool    *v1alpha1.ElasticsearchClusterNodePool
		statefulSet *apps.StatefulSet
		replicaDiff int32
		canScale    bool
		err         bool
	}
	tests := map[string]testT{
		"can scale down statefulset with just 1 empty pilot": {
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 10),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 0),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			replicaDiff: -1,
			canScale:    true,
		},
		"cannot scale down statefulset with a non empty pilot": {
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 100),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			replicaDiff: -1,
			err:         false,
			canScale:    false,
		},
		"cannot scale down node pool when a pilot has not reported its document count": {
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 0),
				pilotWithNameOwner("es-test-data-2", "test", "data"),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 2, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			replicaDiff: -1,
			err:         false,
			canScale:    false,
		},
		"can always scale up": {
			navObjects: []runtime.Object{
				pilotWithNameDocuments("es-test-data-0", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-1", "test", "data", 0),
				pilotWithNameDocuments("es-test-data-2", "test", "data", 100),
			},
			cluster:     clusterWithNodePools("test", nodePoolWithNameReplicasRoles("data", 4, v1alpha1.ElasticsearchRoleData)),
			nodePool:    nodePoolPtrWithNameReplicasRoles("data", 4, v1alpha1.ElasticsearchRoleData),
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test-data", Replicas: int32Ptr(3)}),
			replicaDiff: 1,
			canScale:    true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			kubeObjects := test.kubeObjects
			if test.statefulSet != nil {
				kubeObjects = append(kubeObjects, test.statefulSet)
			}
			navObjects := test.navObjects
			if test.cluster != nil {
				navObjects = append(navObjects, test.cluster)
			}
			fixture := &framework.StateFixture{
				T:                t,
				KubeObjects:      kubeObjects,
				NavigatorObjects: navObjects,
			}
			fixture.Start()
			state := fixture.State()
			scale := &Scale{
				Cluster:  test.cluster,
				NodePool: test.nodePool,
				Replicas: int32(test.nodePool.Replicas) + test.replicaDiff,
			}
			canScale, err := scale.canScaleNodePool(state, test.statefulSet, test.replicaDiff)
			if err != nil && !test.err {
				t.Errorf("Expected no error but got: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			if canScale != test.canScale {
				t.Errorf("Expected %t but got %t", test.canScale, canScale)
			}
		})
	}
}

func TestDeterminePilotsToBeRemoved(t *testing.T) {
	type testT struct {
		inputList      []*v1alpha1.Pilot
		statefulSet    *apps.StatefulSet
		replicaDiff    int32
		expectedOutput []*v1alpha1.Pilot
		err            bool
	}
	tests := []testT{
		{
			inputList:      []*v1alpha1.Pilot{pilotWithName("es-test-0"), pilotWithName("es-test-1"), pilotWithName("es-test-2")},
			statefulSet:    generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test", Replicas: int32Ptr(3)}),
			replicaDiff:    -1,
			expectedOutput: []*v1alpha1.Pilot{pilotWithName("es-test-2")},
			err:            false,
		},
		{
			inputList:      []*v1alpha1.Pilot{pilotWithName("es-test-0"), pilotWithName("es-mixed-1"), pilotWithName("es-test-1")},
			statefulSet:    generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test", Replicas: int32Ptr(2)}),
			replicaDiff:    -1,
			expectedOutput: []*v1alpha1.Pilot{pilotWithName("es-test-1")},
			err:            false,
		},
		{
			inputList:   []*v1alpha1.Pilot{pilotWithName("es-test-0")},
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test", Replicas: int32Ptr(2)}),
			replicaDiff: 0,
			err:         false,
		},
		{
			inputList:   []*v1alpha1.Pilot{pilotWithName("es-test-0")},
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test", Replicas: int32Ptr(1)}),
			replicaDiff: 0,
			err:         false,
		},
		{
			inputList:      []*v1alpha1.Pilot{pilotWithName("es-test-0"), pilotWithName("es-test-1")},
			statefulSet:    generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test", Replicas: int32Ptr(2)}),
			replicaDiff:    -1,
			expectedOutput: []*v1alpha1.Pilot{pilotWithName("es-test-1")},
			err:            false,
		},
		{
			inputList:   []*v1alpha1.Pilot{pilotWithName("es-test-0"), pilotWithName("es-test-1")},
			statefulSet: generate.StatefulSet(generate.StatefulSetConfig{Name: "es-test", Replicas: int32Ptr(2)}),
			replicaDiff: 1,
			err:         false,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			output, err := determinePilotsToBeRemoved(test.inputList, test.statefulSet, test.replicaDiff)
			if err != nil && !test.err {
				t.Errorf("Expected no error but got: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			if len(output) != len(test.expectedOutput) ||
				!reflect.DeepEqual(output, test.expectedOutput) {
				t.Errorf("Expected did not equal actual: %s", pretty.Diff(output, test.expectedOutput))
			}
		})
	}
}

func pilotWithName(name string) *v1alpha1.Pilot {
	return &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels:    map[string]string{},
		},
	}
}

func pilotWithNameOwner(name, clusterName, nodePoolName string) *v1alpha1.Pilot {
	p := pilotWithName(name)
	if p.Labels == nil {
		p.Labels = map[string]string{}
	}
	p.Labels[v1alpha1.ElasticsearchClusterNameLabel] = clusterName
	p.Labels[v1alpha1.ElasticsearchNodePoolNameLabel] = nodePoolName
	return p
}

func pilotWithNameDocuments(name, clusterName, nodePoolName string, documents int64) *v1alpha1.Pilot {
	p := pilotWithNameOwner(name, clusterName, nodePoolName)
	p.Status.Elasticsearch = &v1alpha1.ElasticsearchPilotStatus{
		Documents: &documents,
	}
	return p
}

func clusterWithVersionNodePools(name string, version string, pools ...v1alpha1.ElasticsearchClusterNodePool) *v1alpha1.ElasticsearchCluster {
	return &v1alpha1.ElasticsearchCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1alpha1.ElasticsearchClusterSpec{
			NavigatorClusterConfig: v1alpha1.NavigatorClusterConfig{
				PilotImage: v1alpha1.ImageSpec{
					Repository: "something",
					Tag:        "latest",
				},
			},
			Version:   *semver.New(version),
			NodePools: pools,
		},
	}
}

func clusterWithNodePools(name string, pools ...v1alpha1.ElasticsearchClusterNodePool) *v1alpha1.ElasticsearchCluster {
	return &v1alpha1.ElasticsearchCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1alpha1.ElasticsearchClusterSpec{
			NavigatorClusterConfig: v1alpha1.NavigatorClusterConfig{
				PilotImage: v1alpha1.ImageSpec{
					Repository: "something",
					Tag:        "latest",
				},
			},
			Version:   *semver.New("5.6.2"),
			NodePools: pools,
		},
	}
}

func nodePoolWithNameReplicasRoles(name string, replicas int32, roles ...v1alpha1.ElasticsearchClusterRole) v1alpha1.ElasticsearchClusterNodePool {
	return v1alpha1.ElasticsearchClusterNodePool{
		Name:     name,
		Replicas: replicas,
		Roles:    roles,
	}
}

func nodePoolPtrWithNameReplicasRoles(name string, replicas int32, roles ...v1alpha1.ElasticsearchClusterRole) *v1alpha1.ElasticsearchClusterNodePool {
	obj := nodePoolWithNameReplicasRoles(name, replicas, roles...)
	return &obj
}
