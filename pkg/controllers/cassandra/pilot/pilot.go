package pilot

import (
	"fmt"
	"hash/fnv"
	"reflect"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	hashutil "github.com/jetstack/navigator/pkg/util/hash"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
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
	existingPilot = existingPilot.DeepCopy()
	existingPilot.Status = v1alpha1.PilotStatus{}
	desiredPilot = existingPilot.DeepCopy()
	desiredPilot = updatePilotForCluster(cluster, pod, desiredPilot)
	if !reflect.DeepEqual(desiredPilot, existingPilot) {
		_, err = client.Update(desiredPilot)
	}
	return err
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
	// TODO: Housekeeping. Remove pilots that don't have a corresponding pod.
	return nil
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) *v1alpha1.Pilot {
	pilot := &v1alpha1.Pilot{}
	pilot.SetOwnerReferences(
		[]metav1.OwnerReference{
			util.NewControllerRef(cluster),
		},
	)
	return updatePilotForCluster(cluster, pod, pilot)
}

func updatePilotForCluster(
	cluster *v1alpha1.CassandraCluster,
	pod *v1.Pod,
	pilot *v1alpha1.Pilot,
) *v1alpha1.Pilot {
	pilot.SetName(pod.GetName())
	pilot.SetNamespace(cluster.GetNamespace())
	labels := pilot.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for key, val := range util.ClusterLabels(cluster) {
		labels[key] = val
	}
	pilot.SetLabels(labels)
	ComputeHashAndUpdateAnnotation(pilot)
	return pilot
}

func ComputeHash(p *v1alpha1.Pilot) uint32 {
	hashVar := []interface{}{
		p.Spec,
		p.ObjectMeta,
		p.Labels,
	}
	hasher := fnv.New32()
	hashutil.DeepHashObject(hasher, hashVar)
	return hasher.Sum32()
}

func UpdateHashAnnotation(p *v1alpha1.Pilot, hash uint32) {
	annotations := p.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[HashAnnotationKey] = fmt.Sprintf("%d", hash)
	p.SetAnnotations(annotations)
}

func ComputeHashAndUpdateAnnotation(p *v1alpha1.Pilot) {
	hash := ComputeHash(p)
	UpdateHashAnnotation(p, hash)
}
