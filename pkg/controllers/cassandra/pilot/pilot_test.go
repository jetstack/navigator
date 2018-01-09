package pilot_test

import (
	"testing"

	"k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func nonClusterPod(cluster *v1alpha1.CassandraCluster, name string) *v1.Pod {
	p := clusterPod(cluster, name)
	p.SetOwnerReferences([]metav1.OwnerReference{})
	return p
}

func TestPilotSync(t *testing.T) {
	t.Run(
		"each cluster pod gets a pilot",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.AddObjectK(clusterPod(f.Cluster, "foo"))
			f.AddObjectK(clusterPod(f.Cluster, "bar"))
			f.Run()
			f.AssertPilotsLength(2)
		},
	)
	t.Run(
		"non-cluster pods are ignored",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.AddObjectK(clusterPod(f.Cluster, "foo"))
			f.AddObjectK(nonClusterPod(f.Cluster, "bar"))
			f.Run()
			f.AssertPilotsLength(1)
		},
	)
	t.Run(
		"pilot exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			pod := clusterPod(f.Cluster, "foo")
			pilot := pilot.PilotForCluster(f.Cluster, pod)
			f.AddObjectK(pod)
			f.AddObjectN(pilot)
			f.Run()
			f.AssertPilotsLength(1)
		},
	)
	t.Run(
		"foreign owned pilot",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			pod := clusterPod(f.Cluster, "foo")
			pilot := pilot.PilotForCluster(f.Cluster, pod)
			pilot.SetOwnerReferences([]metav1.OwnerReference{})
			f.AddObjectK(pod)
			f.AddObjectN(pilot)
			f.RunExpectError()
			f.AssertPilotsLength(1)
		},
	)
	t.Run(
		"pilot sync when hash changes",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			pod := clusterPod(f.Cluster, "foo")
			unsyncedPilot := pilot.PilotForCluster(f.Cluster, pod)
			pilot.UpdateHashAnnotation(unsyncedPilot, 0)
			f.AddObjectK(pod)
			f.AddObjectN(unsyncedPilot)
			f.Run()
			f.AssertPilotsLength(1)
			updatedPilot := f.Pilots().Items[0]
			updatedPilotAnnotations := updatedPilot.GetAnnotations()
			hash, ok := updatedPilotAnnotations[pilot.HashAnnotationKey]
			if !ok {
				t.Log(updatedPilotAnnotations)
				t.Error("pilot hash annotation not found")
			}
			if hash == "0" {
				t.Log(updatedPilot)
				t.Error("Pilot was not updated")
			}
		},
	)
	t.Run(
		"pilot no sync if hash matches",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			pod := clusterPod(f.Cluster, "foo")
			// Remove the labels
			unsyncedPilot := pilot.PilotForCluster(f.Cluster, pod)
			unsyncedPilot.SetLabels(map[string]string{})
			pilot.ComputeHashAndUpdateAnnotation(unsyncedPilot)
			f.AddObjectK(pod)
			f.AddObjectN(unsyncedPilot)
			f.Run()
			f.AssertPilotsLength(1)
			updatedPilot := f.Pilots().Items[0]
			updatedLabels := updatedPilot.GetLabels()
			if len(updatedLabels) == 0 {
				t.Log(updatedPilot)
				t.Error("pilot was not updated")
			}
		},
	)
	t.Run(
		"don't clobber custom labels",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			pod := clusterPod(f.Cluster, "foo")
			// Remove the labels
			unsyncedPilot := pilot.PilotForCluster(f.Cluster, pod)
			unsyncedPilot.Labels["foo"] = "bar"
			f.AddObjectK(pod)
			f.AddObjectN(unsyncedPilot)
			f.Run()
			f.AssertPilotsLength(1)
			updatedPilot := f.Pilots().Items[0]
			updatedLabels := updatedPilot.GetLabels()
			if updatedLabels["foo"] != "bar" {
				t.Log(updatedLabels)
				t.Error("custom labels were altered")
			}
		},
	)
}
