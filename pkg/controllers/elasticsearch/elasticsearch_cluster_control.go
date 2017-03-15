package elasticsearch

import (
	"fmt"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/record"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
	"gitlab.jetstack.net/marshal/colonel/pkg/util/errors"
)

type ElasticsearchClusterControl interface {
	SyncElasticsearchCluster(*v1.ElasticsearchCluster) error
}

type defaultElasticsearchClusterControl struct {
	statefulSetLister appslisters.StatefulSetLister
	deploymentLister  extensionslisters.DeploymentLister

	nodePoolControl         ElasticsearchClusterNodePoolControl
	statefulNodePoolControl ElasticsearchClusterNodePoolControl

	recorder record.EventRecorder
}

var _ ElasticsearchClusterControl = &defaultElasticsearchClusterControl{}

func NewElasticsearchClusterControl(
	statefulSetLister appslisters.StatefulSetLister,
	deploymentLister extensionslisters.DeploymentLister,
	nodePoolControl ElasticsearchClusterNodePoolControl,
	statefulNodePoolControl ElasticsearchClusterNodePoolControl,
	recorder record.EventRecorder,
) ElasticsearchClusterControl {
	return &defaultElasticsearchClusterControl{
		statefulSetLister:       statefulSetLister,
		deploymentLister:        deploymentLister,
		nodePoolControl:         nodePoolControl,
		statefulNodePoolControl: statefulNodePoolControl,
		recorder:                recorder,
	}
}

func (e *defaultElasticsearchClusterControl) SyncElasticsearchCluster(
	c *v1.ElasticsearchCluster,
) error {
	for _, np := range c.Spec.NodePools {
		exists, needsUpdate, err := e.nodePoolNeedsUpdate(c, np)

		if err != nil {
			e.recordClusterEvent("sync", c, err)

			if errors.IsTransient(err) {
				return err
			}

			return nil
		}

		if c.DeletionTimestamp != nil {
			needsUpdate = true
		}

		if !needsUpdate {
			continue
		}

		nodePoolUpdater := e.nodePoolUpdater(np)

		switch {
		case c.DeletionTimestamp != nil && exists:
			err = nodePoolUpdater.DeleteElasticsearchClusterNodePool(c, np)
			break
		case exists:
			err = nodePoolUpdater.UpdateElasticsearchClusterNodePool(c, np)
			break
		default:
			err = nodePoolUpdater.CreateElasticsearchClusterNodePool(c, np)
		}

		if err != nil {
			if errors.IsTransient(err) {
				return err
			}

			return nil
		}
	}
	return nil
}

func (e *defaultElasticsearchClusterControl) nodePoolUpdater(np *v1.ElasticsearchClusterNodePool) ElasticsearchClusterNodePoolControl {
	if nodePoolIsStateful(np) {
		return e.statefulNodePoolControl
	}
	return e.nodePoolControl
}

func (e *defaultElasticsearchClusterControl) nodePoolNeedsUpdate(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (exists, needsUpdate bool, err error) {
	if nodePoolIsStateful(np) {
		return e.statefulNodePoolNeedsUpdate(c, np)
	}
	return e.deploymentNodePoolNeedsUpdate(c, np)
}

var notFoundErr = fmt.Errorf("resource not found")

func (e *defaultElasticsearchClusterControl) deploymentForNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (*extensions.Deployment, error) {
	if nodePoolIsStateful(np) {
		return nil, fmt.Errorf("node pool is stateful, but deploymentForNodePool called")
	}

	depls, err := e.deploymentLister.Deployments(c.Namespace).List(labels.Everything())

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Transient(fmt.Errorf("error getting deployments from apiserver: %s", err.Error()))
	}

	for _, depl := range depls {
		if !isManagedByCluster(c, depl.ObjectMeta) {
			continue
		}

		// TODO: switch this to use UIDs set as annotations on the ElasticsearchCluster?
		if depl.Name == nodePoolResourceName(c, np) {
			return depl, nil
		}
	}

	return nil, notFoundErr
}

func (e *defaultElasticsearchClusterControl) deploymentNodePoolNeedsUpdate(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (exists, needsUpdate bool, err error) {
	if nodePoolIsStateful(np) {
		return false, false, fmt.Errorf("node pool is stateful, but deploymentNodePoolNeedsUpdate called")
	}

	depl, err := e.deploymentForNodePool(c, np)

	if err != nil {
		if err == notFoundErr {
			return false, true, nil
		}

		return false, false, errors.Transient(fmt.Errorf("error checking for deployments for node pool '%s'", np.Name))
	}

	// if the desired number of replicas is not equal to the actual
	if *depl.Spec.Replicas != int32(np.Replicas) {
		return true, true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if nodePoolVersionAnnotation(depl.Annotations) != c.Spec.Version {
		return true, true, nil
	}

	return true, false, nil
}

func (e *defaultElasticsearchClusterControl) statefulNodePoolNeedsUpdate(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (exists, needsUpdate bool, err error) {
	if !nodePoolIsStateful(np) {
		return false, false, fmt.Errorf("node pool is not stateful, but statefulNodePoolNeedsUpdate called")
	}

	nodePoolName := nodePoolResourceName(c, np)
	ss, err := e.statefulSetLister.StatefulSets(c.Namespace).Get(nodePoolName)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, true, nil
		}

		return false, false, errors.Transient(fmt.Errorf("error getting statefulset '%s' from apiserver: %s", nodePoolName, err.Error()))
	}

	// if this statefulset is not marked as managed by the cluster, exit with an error and not performing an update to prevent
	// standing on the cluster administrators toes
	if !isManagedByCluster(c, ss.ObjectMeta) {
		return false, false, errors.Transient(fmt.Errorf("found existing statefulset with name, but it is not owned by this ElasticsearchCluster. not updating!"))
	}

	// if the desired number of replicas is not equal to the actual
	if *ss.Spec.Replicas != int32(np.Replicas) {
		return true, true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if nodePoolVersionAnnotation(ss.Annotations) != c.Spec.Version {
		return true, true, nil
	}

	return false, false, nil
	// container, ok := ss.Spec.Template.Spec.Containers[0]

	// // somehow there are no containers in this Pod - trigger an update
	// if !ok {
	// 	return true, nil
	// }

	// if
}

// recordClusterEvent records an event for verb applied to the ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterControl) recordClusterEvent(verb string, cluster *v1.ElasticsearchCluster, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in ElasticsearchCluster %s successful",
			strings.ToLower(verb), cluster.Name)
		e.recorder.Event(cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), cluster.Name, err)
		e.recorder.Event(cluster, apiv1.EventTypeWarning, reason, message)
	}
}
