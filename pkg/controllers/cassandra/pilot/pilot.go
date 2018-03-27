package pilot

import (
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
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

func (c *pilotControl) clusterPods(cluster *v1alpha1.CassandraCluster) ([]*v1.Pod, error) {
	var clusterPods []*v1.Pod
	allPods, err := c.pods.Pods(cluster.Namespace).List(labels.Everything())
	if err != nil {
		return clusterPods, err
	}
	for _, pod := range allPods {
		podControlledByCluster, err := controllers.PodControlledByCluster(
			cluster,
			pod,
			c.statefulSets,
		)
		if err != nil {
			return clusterPods, err
		}
		if !podControlledByCluster {
			continue
		}
		clusterPods = append(clusterPods, pod)
	}
	return clusterPods, nil
}

func (c *pilotControl) createPilot(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) error {
	desiredPilot := PilotForCluster(cluster, pod)
	client := c.naviClient.NavigatorV1alpha1().Pilots(desiredPilot.GetNamespace())
	lister := c.pilots.Pilots(desiredPilot.GetNamespace())
	existingPilot, err := lister.Get(desiredPilot.GetName())
	// Pilot already exists
	if err == nil {
		return util.OwnerCheck(existingPilot, cluster)
	}
	// The only error we expect is that the pilot does not exist.
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	_, err = client.Create(desiredPilot)
	return err
}

func (c *pilotControl) syncPilots(cluster *v1alpha1.CassandraCluster) error {
	pods, err := c.clusterPods(cluster)
	if err != nil {
		return err
	}
	for _, pod := range pods {
		err = c.createPilot(cluster, pod)
		if err != nil {
			return err
		}
	}
	return err
}

func (c *pilotControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	err := c.syncPilots(cluster)
	if err != nil {
		return err
	}
	// TODO: Housekeeping. Remove pilots that don't have a corresponding pod.
	return nil
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) *v1alpha1.Pilot {
	return &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			Labels:          util.ClusterLabels(cluster),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
	}
}
