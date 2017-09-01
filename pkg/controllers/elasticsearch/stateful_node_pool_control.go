package elasticsearch

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

// statefulElasticsearchClusterNodePoolControl manages the lifecycle of a
// stateful node pool. It can be used to create, update and delete pools.
//
// This is an implementation of the ElasticsearchClusterNodePoolControl interface
// as defined in interfaces.go.
type statefulElasticsearchClusterNodePoolControl struct {
	kubeClient        *kubernetes.Clientset
	statefulSetLister appslisters.StatefulSetLister

	recorder record.EventRecorder
}

var _ ElasticsearchClusterNodePoolControl = &statefulElasticsearchClusterNodePoolControl{}

func NewStatefulElasticsearchClusterNodePoolControl(
	kubeClient *kubernetes.Clientset,
	statefulSetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) ElasticsearchClusterNodePoolControl {
	return &statefulElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
	}
}

func (e *statefulElasticsearchClusterNodePoolControl) CreateElasticsearchClusterNodePool(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) error {
	ss, err := nodePoolStatefulSet(c, np)

	if err != nil {
		e.recordNodePoolEvent("create", c, np, err)
		return fmt.Errorf("error generating statefulset manifest: %s", err.Error())
	}

	ss, err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Create(ss)

	if err != nil {
		e.recordNodePoolEvent("create", c, np, err)
		return fmt.Errorf("error creating statefulset: %s", err.Error())
	}

	e.recordNodePoolEvent("create", c, np, err)
	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) UpdateElasticsearchClusterNodePool(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) error {
	ss, err := nodePoolStatefulSet(c, np)

	if err != nil {
		e.recordNodePoolEvent("update", c, np, err)
		return fmt.Errorf("error generating statefulset manifest: %s", err.Error())
	}

	ss, err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(ss)

	if err != nil {
		e.recordNodePoolEvent("update", c, np, err)
		return fmt.Errorf("error updating statefulset: %s", err.Error())
	}

	e.recordNodePoolEvent("update", c, np, err)
	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) DeleteElasticsearchClusterNodePool(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) error {
	ss, err := nodePoolStatefulSet(c, np)

	if err != nil {
		e.recordNodePoolEvent("delete", c, np, err)
		return fmt.Errorf("error generating statefulset manifest: %s", err.Error())
	}

	err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Delete(ss.Name, &metav1.DeleteOptions{OrphanDependents: &falseVar})

	if err != nil {
		e.recordNodePoolEvent("delete", c, np, err)
		return fmt.Errorf("error deleting statefulset: %s", err.Error())
	}

	e.recordNodePoolEvent("delete", c, np, err)
	return nil
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *statefulElasticsearchClusterNodePoolControl) recordNodePoolEvent(verb string, cluster v1alpha1.ElasticsearchCluster, pool v1alpha1.ElasticsearchClusterNodePool, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s StatefulNodePool %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), pool.Name, cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s StatefulNodePool %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), pool.Name, cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
