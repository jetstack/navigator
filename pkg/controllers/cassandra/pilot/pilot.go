package pilot

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
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

func (c *pilotControl) createOrUpdatePilot(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) error {
	desiredPilot := PilotForCluster(cluster, pod)
	client := c.naviClient.NavigatorV1alpha1().Pilots(desiredPilot.GetNamespace())
	lister := c.pilots.Pilots(desiredPilot.GetNamespace())
	existingPilot, err := lister.Get(desiredPilot.GetName())
	if k8sErrors.IsNotFound(err) {
		_, err = client.Create(desiredPilot)
		return err
	}
	if err != nil {
		return err
	}
	err = util.OwnerCheck(existingPilot, cluster)
	if err != nil {
		return err
	}
	desiredPilot = existingPilot.DeepCopy()
	updatePilotForCluster(cluster, pod, desiredPilot)
	_, err = client.Update(desiredPilot)
	return err
}

func (c *pilotControl) removeUnusedPilots(
	cluster *v1alpha1.CassandraCluster,
) error {
	expectedPilotNames := map[string]bool{}
	clusterPods, err := c.clusterPods(cluster)
	if err != nil {
		return err
	}
	for _, pod := range clusterPods {
		expectedPilotNames[pod.Name] = true
	}
	existingPilots, err := c.pilots.Pilots(cluster.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	client := c.naviClient.NavigatorV1alpha1().Pilots(cluster.Namespace)
	for _, pilot := range existingPilots {
		err := util.OwnerCheck(pilot, cluster)
		if err != nil {
			return err
		}
		_, found := expectedPilotNames[pilot.Name]
		if !found {
			err := client.Delete(pilot.Name, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *pilotControl) syncPilots(cluster *v1alpha1.CassandraCluster) error {
	pods, err := c.clusterPods(cluster)
	if err != nil {
		return err
	}
	for _, pod := range pods {
		err = c.createOrUpdatePilot(cluster, pod)
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
	err = c.removeUnusedPilots(cluster)
	return err
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) *v1alpha1.Pilot {
	pilot := &v1alpha1.Pilot{}
	ownerRefs := pilot.GetOwnerReferences()
	ownerRefs = append(ownerRefs, util.NewControllerRef(cluster))
	pilot.SetOwnerReferences(ownerRefs)
	return updatePilotForCluster(cluster, pod, pilot)
}

func updatePilotForCluster(
	cluster *v1alpha1.CassandraCluster,
	pod *v1.Pod,
	pilot *v1alpha1.Pilot,
) *v1alpha1.Pilot {
	pilot.SetName(pod.GetName())
	pilot.SetNamespace(cluster.GetNamespace())
	pilot.SetLabels(util.ClusterLabels(cluster))
	return pilot
}
