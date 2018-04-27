package pilot

import (
	"fmt"

	"k8s.io/api/apps/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

const (
	HashAnnotationKey = "navigator.jetstack.io/cassandra-pilot-hash"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type pilotControl struct {
	naviClient   navigator.Interface
	pilots       navlisters.PilotLister
	pods         corelisters.PodLister
	statefulSets appslisters.StatefulSetLister
	recorder     record.EventRecorder
}

var _ Interface = &pilotControl{}

func NewControl(
	naviClient navigator.Interface,
	pilots navlisters.PilotLister,
	pods corelisters.PodLister,
	statefulSets appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) *pilotControl {
	return &pilotControl{
		naviClient:   naviClient,
		pilots:       pilots,
		pods:         pods,
		statefulSets: statefulSets,
		recorder:     recorder,
	}

}

func (c *pilotControl) pilotsForSet(
	cluster *v1alpha1.CassandraCluster,
	ss *v1beta1.StatefulSet,
) []*v1alpha1.Pilot {
	pilots := make([]*v1alpha1.Pilot, *ss.Spec.Replicas)
	for i := int32(0); i < *ss.Spec.Replicas; i++ {
		pilots[i] = PilotForCluster(cluster, ss, i)
	}
	return pilots
}

func (c *pilotControl) createPilot(pilot *v1alpha1.Pilot) error {
	_, err := c.naviClient.NavigatorV1alpha1().Pilots(pilot.Namespace).Create(pilot)
	if k8sErrors.IsAlreadyExists(err) {
		glog.Warning("Pilot already exists %s/%s.", pilot.Namespace, pilot.Name)
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "unable to create pilot")
	}
	glog.V(4).Infof("Created pilot %s/%s.", pilot.Namespace, pilot.Name)
	return nil
}

func (c *pilotControl) deletePilot(cluster *v1alpha1.CassandraCluster, pilot *v1alpha1.Pilot) error {
	err := util.OwnerCheck(pilot, cluster)
	if err != nil {
		glog.Errorf(
			"Skipping deletion of foreign owned pilot %s/%s: %s.",
			pilot.Namespace, pilot.Name, err,
		)
		return nil
	}

	_, err = c.pods.Pods(cluster.Namespace).Get(pilot.Name)
	if err == nil {
		glog.V(4).Infof(
			"Skipping deletion of pilot %s/%s because its pod still exists.",
			pilot.Namespace, pilot.Name,
		)
		return nil
	}
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unable to get pod for pilot")
	}
	err = c.naviClient.NavigatorV1alpha1().
		Pilots(cluster.Namespace).Delete(pilot.Name, &metav1.DeleteOptions{})
	if k8sErrors.IsNotFound(err) {
		glog.Warning("Pilot already deleted %s/%s.", pilot.Namespace, pilot.Name)
		return nil
	}
	if err != nil {
		return errors.Wrapf(
			err, "unable to delete pilot %s/%s", pilot.Namespace, pilot.Name,
		)
	}
	glog.V(4).Infof("Deleted pilot %s/%s.", pilot.Namespace, pilot.Name)
	return nil
}

// Sync ensures the correct number of Pilots for each nodepool.
//
// For each nodepool StatefulSet:
// * Create a number of pilots to match the number of pods that will be created for the StatefulSet.
// * Delete higher index pilots which have been left behind after the statefulset has been scaled in.
// * Do not delete a pilot if there is a pod with a matching name,
//   (that pod won't be able to decommission its self
//    unless it can read its desired configuration from its pilot)
// * Do not delete a pilot unless it is owned by the cluster that is being synchronised.
//   (this is not an expected state,
//    but we don't want to delete anything unless it was created by us)
func (c *pilotControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	selector, err := util.SelectorForClusterNodePools(cluster)
	if err != nil {
		return errors.Wrap(err, "unable to create cluster nodepools selector")
	}
	statefulSets, err := c.statefulSets.StatefulSets(cluster.Namespace).List(selector)
	if err != nil {
		return errors.Wrap(err, "unable to list statefulsets")
	}

	actualPilots, err := c.pilots.Pilots(cluster.Namespace).List(selector)
	if err != nil {
		return errors.Wrap(err, "unable to list pilots")
	}
	actualPilotNames := setOfPilotNames(actualPilots)

	expectedPilots := []*v1alpha1.Pilot{}
	for _, set := range statefulSets {
		expectedPilots = append(expectedPilots, c.pilotsForSet(cluster, set)...)
	}
	expectedPilotNames := setOfPilotNames(expectedPilots)

	pilotsToCreate := expectedPilotNames.Difference(actualPilotNames)
	glog.V(4).Infof("Creating pilots: %v", pilotsToCreate.List())
	for _, pilot := range expectedPilots {
		if !pilotsToCreate.Has(pilot.Name) {
			continue
		}
		err := c.createPilot(pilot)
		if err != nil {
			return errors.Wrap(err, "error in createPilot")
		}
	}

	pilotsToDelete := actualPilotNames.Difference(expectedPilotNames)
	glog.V(4).Infof("Deleting pilots: %v", pilotsToDelete.List())
	for _, pilot := range actualPilots {
		if !pilotsToDelete.Has(pilot.Name) {
			continue
		}
		err := c.deletePilot(cluster, pilot)
		if err != nil {
			return errors.Wrap(err, "error in deletePilot")
		}
	}
	return nil
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, ss *v1beta1.StatefulSet, index int32) *v1alpha1.Pilot {
	o := &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%d", ss.Name, index),
			Namespace:       ss.Namespace,
			Labels:          util.ClusterLabels(cluster),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
	}
	o.Labels[v1alpha1.CassandraNodePoolNameLabel] = ss.Labels[v1alpha1.CassandraNodePoolNameLabel]
	o.Labels[v1alpha1.CassandraNodePoolIndexLabel] = fmt.Sprintf("%d", index)
	return o
}

func setOfPilotNames(pilots []*v1alpha1.Pilot) sets.String {
	names := sets.NewString()
	for _, pilot := range pilots {
		names.Insert(pilot.Name)
	}
	return names
}
