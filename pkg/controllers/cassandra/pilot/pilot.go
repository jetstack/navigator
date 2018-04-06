package pilot

import (
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/golang/glog"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/cassandra/version"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
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

func (c *pilotControl) updateDiscoveredVersions(cluster *v1alpha1.CassandraCluster) error {
	glog.V(4).Infof("updateDiscoveredVersions for cluster: %s", cluster.Name)
	selector, err := util.SelectorForCluster(cluster)
	if err != nil {
		return err
	}
	pilots, err := c.pilots.List(selector)
	if err != nil {
		return err
	}
	if len(pilots) < 1 {
		glog.V(4).Infof("No pilots found matching selector: %s", selector)
	}
	for _, pilot := range pilots {
		nodePoolNameForPilot, nodePoolNameFound := pilot.Labels[util.NodePoolNameLabelKey]
		if !nodePoolNameFound {
			glog.Warningf("Skipping pilot without NodePoolNameLabelKey: %s", pilot.Name)
			continue
		}
		nodePoolStatus := cluster.Status.NodePools[nodePoolNameForPilot]
		switch {
		case pilot.Status.Cassandra == nil:
			glog.V(4).Infof(
				"Pilot %s/%s has no status. Setting nodepool version to nil",
				pilot.Namespace, pilot.Name,
			)
			nodePoolStatus.Version = nil
		case pilot.Status.Cassandra.Version == nil:
			glog.V(4).Infof(
				"Pilot %s/%s has not reported its version. Setting nodepool version to nil",
				pilot.Namespace, pilot.Name,
			)
			nodePoolStatus.Version = nil
		case nodePoolStatus.Version == nil:
			nodePoolStatus.Version = nil
		case pilot.Status.Cassandra.Version.LessThan(nodePoolStatus.Version):
			glog.V(4).Infof(
				"Found lower pilot version: %s, %s",
				nodePoolNameForPilot, pilotVersionStatus,
			)
			nodePoolStatus.Version = pilotVersionStatus
		}
		cluster.Status.NodePools[nodePoolNameForPilot] = nodePoolStatus
	}
	return nil
}

func (c *pilotControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	err := c.syncPilots(cluster)
	if err != nil {
		return err
	}
	// TODO: Housekeeping. Remove pilots that don't have a corresponding pod.

	return c.updateDiscoveredVersions(cluster)
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) *v1alpha1.Pilot {
	o := &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			Labels:          util.ClusterLabels(cluster),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
	}
	o.Labels[util.NodePoolNameLabelKey] = pod.Labels[util.NodePoolNameLabelKey]
	return o
}

func UpdateLabels(
	o metav1.Object,
	newLabels map[string]string,
) {
	labels := o.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for key, val := range newLabels {
		labels[key] = val
	}
	o.SetLabels(labels)
}

type PilotBuilder struct {
	pilot *v1alpha1.Pilot
}

func NewPilotBuilder() *PilotBuilder {
	return &PilotBuilder{
		pilot: &v1alpha1.Pilot{},
	}
}

func (pb *PilotBuilder) ForCluster(cluster metav1.Object) *PilotBuilder {
	UpdateLabels(pb.pilot, util.ClusterLabels(cluster))
	pb.pilot.SetNamespace(cluster.GetNamespace())
	pb.pilot.SetOwnerReferences(
		append(
			pb.pilot.GetOwnerReferences(),
			util.NewControllerRef(cluster),
		),
	)
	return pb
}

func (pb *PilotBuilder) ForNodePool(np *v1alpha1.CassandraClusterNodePool) *PilotBuilder {
	UpdateLabels(
		pb.pilot,
		map[string]string{
			util.NodePoolNameLabelKey: np.Name,
		},
	)
	return pb
}

func (pb *PilotBuilder) WithCassandraStatus() *PilotBuilder {
	pb.pilot.Status.Cassandra = &v1alpha1.CassandraPilotStatus{}
	return pb
}

func (pb *PilotBuilder) WithDiscoveredCassandraVersion(v string) *PilotBuilder {
	pb.WithCassandraStatus()
	pb.pilot.Status.Cassandra.Version = version.New(v)
	return pb
}

func (pb *PilotBuilder) Build() *v1alpha1.Pilot {
	return pb.pilot
}
