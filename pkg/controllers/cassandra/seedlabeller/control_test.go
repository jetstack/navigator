package seedlabeller_test

import (
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/seedlabeller"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
	"github.com/jetstack/navigator/pkg/util"
)

func CheckSeedLabel(podName, seedLabelValue string, podNamespace string, t *testing.T, state *controllers.State) {
	p, err := state.Clientset.
		CoreV1().
		Pods(podNamespace).
		Get(podName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Labels[service.SeedLabelKey] != seedLabelValue {
		t.Errorf("unexpected seed label: %s", p.Labels)
	}
}

func TestSeedLabellerSync(t *testing.T) {
	cluster := casstesting.ClusterForTest()
	np0 := &cluster.Spec.NodePools[0]
	ss0 := nodepool.StatefulSetForCluster(cluster, np0)
	pod0 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cass-cassandra-1-RingNodes-0",
			Namespace: cluster.Namespace,
		},
	}
	pod0LabelMissing := pod0.DeepCopy()
	pod0LabelMissing.SetLabels(map[string]string{})
	pod0ValueIncorrect := pod0LabelMissing.DeepCopy()
	pod0ValueIncorrect.Labels[service.SeedLabelKey] = "blah"

	clusterOneSeed := cluster.DeepCopy()
	clusterOneSeed.Spec.NodePools[0].Seeds = util.Int64Ptr(1)

	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cass-cassandra-1-RingNodes-1",
			Namespace: cluster.Namespace,
		},
	}

	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cass-cassandra-1-RingNodes-2",
			Namespace: cluster.Namespace,
		},
	}

	type testT struct {
		kubeObjects []runtime.Object
		navObjects  []runtime.Object
		cluster     *v1alpha1.CassandraCluster
		assertions  func(*testing.T, *controllers.State)
		expectErr   bool
	}

	tests := map[string]testT{
		"ignore missing pod": {
			kubeObjects: []runtime.Object{ss0},
			navObjects:  []runtime.Object{cluster},
			cluster:     cluster,
		},
		"add label if nil labels": {
			kubeObjects: []runtime.Object{ss0, pod0},
			navObjects:  []runtime.Object{cluster},
			cluster:     cluster,
			assertions: func(t *testing.T, state *controllers.State) {
				CheckSeedLabel(pod0.Name, seedprovider.SeedLabelValue, pod0.Namespace, t, state)
			},
		},
		"add label if key missing": {
			kubeObjects: []runtime.Object{ss0, pod0LabelMissing},
			navObjects:  []runtime.Object{cluster},
			cluster:     cluster,
			assertions: func(t *testing.T, state *controllers.State) {
				CheckSeedLabel(pod0.Name, seedprovider.SeedLabelValue, pod0.Namespace, t, state)
			},
		},
		"fix label if value incorrect": {
			kubeObjects: []runtime.Object{ss0, pod0ValueIncorrect},
			navObjects:  []runtime.Object{cluster},
			cluster:     cluster,
			assertions: func(t *testing.T, state *controllers.State) {
				CheckSeedLabel(pod0.Name, seedprovider.SeedLabelValue, pod0.Namespace, t, state)
			},
		},
		"add multiple seeds": {
			kubeObjects: []runtime.Object{ss0, pod0, pod1, pod2},
			navObjects:  []runtime.Object{cluster},
			cluster:     cluster,
			assertions: func(t *testing.T, state *controllers.State) {
				CheckSeedLabel(pod1.Name, seedprovider.SeedLabelValue, pod1.Namespace, t, state)
			},
		},
		"don't add too many seeds": {
			kubeObjects: []runtime.Object{ss0, pod0, pod1, pod2},
			navObjects:  []runtime.Object{cluster},
			cluster:     cluster,
			assertions: func(t *testing.T, state *controllers.State) {
				CheckSeedLabel(pod2.Name, "", pod2.Namespace, t, state)
			},
		},
		"delete label if seed number decreased": {
			kubeObjects: []runtime.Object{ss0, pod0, pod1, pod2},
			navObjects:  []runtime.Object{cluster},
			cluster:     clusterOneSeed,
			assertions: func(t *testing.T, state *controllers.State) {
				CheckSeedLabel(pod1.Name, "", pod1.Namespace, t, state)
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
				c := seedlabeller.NewControl(
					state.Clientset,
					state.StatefulSetLister,
					state.PodLister,
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
