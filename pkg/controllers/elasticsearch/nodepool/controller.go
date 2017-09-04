package nodepool

import (
	"fmt"
	"strings"

	apps "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
	"github.com/jetstack-experimental/navigator/pkg/util/errors"
)

type Interface interface {
	Sync(v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error)
}

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

var _ Interface = &statefulElasticsearchClusterNodePoolControl{}

func NewStatefulElasticsearchClusterNodePoolControl(
	kubeClient *kubernetes.Clientset,
	statefulSetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) Interface {
	return &statefulElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
	}
}

func (e *statefulElasticsearchClusterNodePoolControl) Sync(c v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error) {
	for _, np := range c.Spec.NodePools {
		ss, err := util.NodePoolStatefulSet(c, np)

		if err != nil {
			e.recordNodePoolEvent("update", c, np, err)
			return c.Status, fmt.Errorf("error generating statefulset manifest: %s", err.Error())
		}

		ss, err = e.kubeClient.AppsV1beta1().StatefulSets(c.Namespace).Update(ss)

		if err != nil {
			e.recordNodePoolEvent("update", c, np, err)
			return c.Status, fmt.Errorf("error updating statefulset: %s", err.Error())
		}

		e.recordNodePoolEvent("update", c, np, err)
		return c.Status, nil
	}

	return c.Status, nil
}

func (e *statefulElasticsearchClusterNodePoolControl) needsUpdate(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (exists bool, needsUpdate bool, err error) {
	ss, err := e.statefulSetForNodePool(c, np)

	if k8sErrors.IsNotFound(err) {
		return false, true, nil
	}

	if err != nil {
		return false, false, errors.Transient(fmt.Errorf("error checking for statefulsets for node pool '%s'", np.Name))
	}

	if !util.IsManagedByCluster(c, ss.ObjectMeta) {
		return false, false, fmt.Errorf("statefulset '%s' found but not managed by cluster", ss.Name)
	}

	// if the desired number of replicas is not equal to the actual
	if *ss.Spec.Replicas != int32(np.Replicas) {
		return true, true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if util.NodePoolVersionAnnotation(ss.Annotations) != c.Spec.Version {
		return true, true, nil
	}

	if ss.Spec.Template.Spec.Containers[0].Image != c.Spec.Image.Repository+":"+c.Spec.Image.Tag {
		return true, true, nil
	}

	return true, false, nil
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

func (e *statefulElasticsearchClusterNodePoolControl) statefulSetForNodePool(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (*apps.StatefulSet, error) {
	sets, err := e.statefulSetLister.StatefulSets(c.Namespace).List(labels.Everything())

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Transient(fmt.Errorf("error getting list of statefulsets: %s", err.Error()))
		}

		return nil, errors.Transient(fmt.Errorf("error getting statefulsets from apiserver: %s", err.Error()))
	}

	for _, ss := range sets {
		if !util.IsManagedByCluster(c, ss.ObjectMeta) {
			continue
		}

		// TODO: switch this to use UIDs set as annotations on the ElasticsearchCluster?
		if ss.Name == util.NodePoolResourceName(c, np) {
			return ss, nil
		}
	}

	return nil, k8sErrors.NewNotFound(schema.GroupResource{
		Group:    apps.GroupName,
		Resource: "StatefulSet",
	}, "")
}
