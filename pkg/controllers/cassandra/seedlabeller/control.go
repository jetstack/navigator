package seedlabeller

import (
	"fmt"

	"github.com/golang/glog"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultSeedLabeller struct {
	kubeClient        kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	pods              corelisters.PodLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultSeedLabeller{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	pods corelisters.PodLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultSeedLabeller{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		pods:              pods,
		recorder:          recorder,
	}
}

func (c *defaultSeedLabeller) labelSeedNodes(
	cluster *v1alpha1.CassandraCluster,
	set *appsv1beta1.StatefulSet,
) error {
	// TODO: make number of seed nodes configurable
	pod, err := c.pods.Pods(cluster.Namespace).Get(fmt.Sprintf("%s-%d", set.Name, 0))
	if err != nil {
		glog.Warningf("Couldn't get stateful set pod: %v", err)
		return nil
	}
	labels := pod.Labels
	value := labels[service.SeedLabelKey]
	if value == service.SeedLabelValue {
		return nil
	}
	if labels == nil {
		labels = map[string]string{}
	}
	labels[service.SeedLabelKey] = service.SeedLabelValue
	podCopy := pod.DeepCopy()
	podCopy.SetLabels(labels)
	_, err = c.kubeClient.CoreV1().Pods(podCopy.Namespace).Update(podCopy)
	return err
}

func (c *defaultSeedLabeller) Sync(cluster *v1alpha1.CassandraCluster) error {
	sets, err := util.StatefulSetsForCluster(cluster, c.statefulSetLister)
	if err != nil {
		return err
	}
	for _, s := range sets {
		err = c.labelSeedNodes(cluster, s)
		if err != nil {
			return err
		}
	}
	return nil
}
