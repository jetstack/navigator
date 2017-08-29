package couchbase

import (
	"fmt"
	"strings"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/util/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/apis/extensions"
	"k8s.io/client-go/tools/record"
)

var notFoundErr = fmt.Errorf("resource not found")

type defaultCouchbaseClusterControl struct {
	kubeClient *kubernetes.Clientset

	statefulSetLister    appslisters.StatefulSetLister
	deploymentLister     extensionslisters.DeploymentLister
	serviceAccountLister corelisters.ServiceAccountLister
	serviceLister        corelisters.ServiceLister

	nodePoolControl         CouchbaseClusterNodePoolControl
	statefulNodePoolControl CouchbaseClusterNodePoolControl
	serviceAccountControl   CouchbaseClusterServiceAccountControl
	clientServiceControl    CouchbaseClusterServiceControl
	discoveryServiceControl CouchbaseClusterServiceControl

	recorder record.EventRecorder
}

var _ CouchbaseClusterControl = &defaultCouchbaseClusterControl{}

func NewCouchbaseClusterControl(
	statefulSetLister appslisters.StatefulSetLister,
	deploymentLister extensionslisters.DeploymentLister,
	serviceAccountLister corelisters.ServiceAccountLister,
	serviceLister corelisters.ServiceLister,
	nodePoolControl CouchbaseClusterNodePoolControl,
	statefulNodePoolControl CouchbaseClusterNodePoolControl,
	serviceAccountControl CouchbaseClusterServiceAccountControl,
	clientServiceControl CouchbaseClusterServiceControl,
	discoveryServiceControl CouchbaseClusterServiceControl,
	recorder record.EventRecorder,
) CouchbaseClusterControl {
	return &defaultCouchbaseClusterControl{
		statefulSetLister:       statefulSetLister,
		deploymentLister:        deploymentLister,
		serviceAccountLister:    serviceAccountLister,
		serviceLister:           serviceLister,
		nodePoolControl:         nodePoolControl,
		statefulNodePoolControl: statefulNodePoolControl,
		serviceAccountControl:   serviceAccountControl,
		clientServiceControl:    clientServiceControl,
		discoveryServiceControl: discoveryServiceControl,
		recorder:                recorder,
	}
}

func (c *defaultCouchbaseClusterControl) SyncCouchbaseCluster(
	cluster v1alpha1.CouchbaseCluster,
) error {
	var err error

	if err = c.syncServiceAccount(cluster); err != nil {
		c.recordClusterEvent("sync", cluster, err)
		return err
	}

	if err = c.syncService(cluster, c.clientServiceControl); err != nil {
		c.recordClusterEvent("sync", cluster, err)
		return err
	}

	if err = c.syncService(cluster, c.discoveryServiceControl); err != nil {
		c.recordClusterEvent("sync", cluster, err)
		return err
	}

	for _, np := range cluster.Spec.NodePools {
		if err = c.syncNodePool(cluster, np); err != nil {
			c.recordClusterEvent("sync", cluster, err)
			return err
		}
	}

	c.recordClusterEvent("sync", cluster, err)
	return nil
}

func (e *defaultCouchbaseClusterControl) nodePoolNeedsUpdate(c v1alpha1.CouchbaseCluster, np v1alpha1.CouchbaseClusterNodePool) (exists, needsUpdate bool, err error) {
	if nodePoolIsStateful(np) {
		return e.statefulNodePoolNeedsUpdate(c, np)
	}
	return e.deploymentNodePoolNeedsUpdate(c, np)
}

func (e *defaultCouchbaseClusterControl) deploymentNodePoolNeedsUpdate(c v1alpha1.CouchbaseCluster, np v1alpha1.CouchbaseClusterNodePool) (exists, needsUpdate bool, err error) {
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

	if !isManagedByCluster(c, depl.ObjectMeta) {
		return false, false, fmt.Errorf("deployment '%s' found but not managed by cluster", depl.Name)
	}

	// if the desired number of replicas is not equal to the actual
	if *depl.Spec.Replicas != int32(np.Replicas) {
		return true, true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if nodePoolVersionAnnotation(depl.Annotations) != c.Spec.Version {
		return true, true, nil
	}

	if depl.Spec.Template.Spec.Containers[0].Image != c.Spec.Image.Repository+":"+c.Spec.Image.Tag {
		return true, true, nil
	}

	return true, false, nil
}

func (e *defaultCouchbaseClusterControl) deploymentForNodePool(c v1alpha1.CouchbaseCluster, np v1alpha1.CouchbaseClusterNodePool) (*extensions.Deployment, error) {
	if nodePoolIsStateful(np) {
		return nil, fmt.Errorf("node pool is stateful, but deploymentForNodePool called")
	}

	depls, err := e.deploymentLister.Deployments(c.Namespace).List(labels.Everything())

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errors.Transient(fmt.Errorf("error getting list of deployments: %s", err.Error()))
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

func (e *defaultCouchbaseClusterControl) statefulNodePoolNeedsUpdate(c v1alpha1.CouchbaseCluster, np v1alpha1.CouchbaseClusterNodePool) (exists, needsUpdate bool, err error) {
	if !nodePoolIsStateful(np) {
		return false, false, fmt.Errorf("node pool is not stateful, but statefulNodePoolNeedsUpdate called")
	}

	ss, err := e.statefulSetForNodePool(c, np)

	if err != nil {
		if err == notFoundErr {
			return false, true, nil
		}

		return false, false, errors.Transient(fmt.Errorf("error checking for statefulsets for node pool '%s'", np.Name))
	}

	if !isManagedByCluster(c, ss.ObjectMeta) {
		return false, false, fmt.Errorf("statefulset '%s' found but not managed by cluster", ss.Name)
	}

	// if the desired number of replicas is not equal to the actual
	if *ss.Spec.Replicas != int32(np.Replicas) {
		return true, true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if nodePoolVersionAnnotation(ss.Annotations) != c.Spec.Version {
		return true, true, nil
	}

	if ss.Spec.Template.Spec.Containers[0].Image != c.Spec.Image.Repository+":"+c.Spec.Image.Tag {
		return true, true, nil
	}

	return true, false, nil
}

func (e *defaultCouchbaseClusterControl) syncNodePool(c v1alpha1.CouchbaseCluster, np v1alpha1.CouchbaseClusterNodePool) error {
	exists, needsUpdate, err := e.nodePoolNeedsUpdate(c, np)

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	if c.DeletionTimestamp != nil {
		needsUpdate = true
	}

	if !needsUpdate {
		return nil
	}

	nodePoolUpdater := e.nodePoolUpdater(np)

	switch {
	case c.DeletionTimestamp != nil && exists:
		err = nodePoolUpdater.DeleteCouchbaseClusterNodePool(c, np)
		break
	case exists:
		err = nodePoolUpdater.UpdateCouchbaseClusterNodePool(c, np)
		break
	default:
		err = nodePoolUpdater.CreateCouchbaseClusterNodePool(c, np)
	}

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	return nil
}

func (e *defaultCouchbaseClusterControl) syncService(c v1alpha1.CouchbaseCluster, ctrl CouchbaseClusterServiceControl) error {
	exists, needsUpdate, err := e.serviceNeedsUpdate(c, ctrl.NameSuffix())

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	if c.DeletionTimestamp != nil {
		needsUpdate = true
	}

	if !needsUpdate {
		return nil
	}

	switch {
	case c.DeletionTimestamp != nil && exists:
		err = ctrl.DeleteElasticsearchClusterService(c)
		break
	case exists:
		err = ctrl.UpdateElasticsearchClusterService(c)
		break
	default:
		err = ctrl.CreateElasticsearchClusterService(c)
	}

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	return nil
}

func (e *defaultCouchbaseClusterControl) serviceNeedsUpdate(c v1alpha1.CouchbaseCluster, nameSuffix string) (exists, needsUpdate bool, err error) {
	svcName := clusterService(c, nameSuffix, false, nil).Name

	svcs, err := e.serviceLister.Services(c.Namespace).List(labels.Everything())

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, false, errors.Transient(fmt.Errorf("error getting list of services: %s", err.Error()))
		}

		return false, false, errors.Transient(fmt.Errorf("error getting services from apiserver: %s", err.Error()))
	}

	if len(svcs) == 0 {
		return false, true, nil
	}

	for _, svc := range svcs {
		// TODO: switch this to use UIDs set as annotations on the ElasticsearchCluster?
		if svc.Name == svcName {
			if isManagedByCluster(c, svc.ObjectMeta) {
				return true, false, nil
			}
			return false, false, fmt.Errorf("service '%s' found but not managed by cluster", svcName)
		}
	}

	return false, true, nil
}

func (e *defaultCouchbaseClusterControl) syncServiceAccount(cluster v1alpha1.CouchbaseCluster) error {
	exists, needsUpdate, err := e.serviceAccountNeedsUpdate(cluster)

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	if cluster.DeletionTimestamp != nil {
		needsUpdate = true
	}

	if !needsUpdate {
		return nil
	}

	switch {
	case cluster.DeletionTimestamp != nil && exists:
		err = e.serviceAccountControl.DeleteElasticsearchClusterServiceAccount(cluster)
		break
	case exists:
		err = e.serviceAccountControl.UpdateElasticsearchClusterServiceAccount(cluster)
		break
	default:
		err = e.serviceAccountControl.CreateElasticsearchClusterServiceAccount(cluster)
	}

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	return nil
}

// recordClusterEvent records an event for verb applied to the CouchbaseCluster. If err is nil the generated event will
// have a reason of apiv1.EventTypeNormal. If err is not nil the generated event will have a reason of apiv1.EventTypeWarning.
func (e *defaultCouchbaseClusterControl) recordClusterEvent(verb string, cluster v1alpha1.CouchbaseCluster, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in CouchbaseCluster %s successful",
			strings.ToLower(verb), cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in CouchbaseCluster %s failed error: %s",
			strings.ToLower(verb), cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
