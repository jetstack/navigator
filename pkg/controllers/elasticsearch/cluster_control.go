package elasticsearch

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/nodepool"
	"github.com/jetstack-experimental/navigator/pkg/util/errors"
)

type defaultElasticsearchClusterControl struct {
	kubeClient *kubernetes.Clientset

	statefulSetLister    appslisters.StatefulSetLister
	deploymentLister     extensionslisters.DeploymentLister
	serviceAccountLister corelisters.ServiceAccountLister
	serviceLister        corelisters.ServiceLister

	nodePoolController      nodepool.Interface
	statefulNodePoolControl ElasticsearchClusterNodePoolControl
	serviceAccountControl   ElasticsearchClusterServiceAccountControl
	clientServiceControl    ElasticsearchClusterServiceControl
	discoveryServiceControl ElasticsearchClusterServiceControl

	recorder record.EventRecorder
}

var _ ElasticsearchClusterControl = &defaultElasticsearchClusterControl{}

func NewElasticsearchClusterControl(
	statefulSetLister appslisters.StatefulSetLister,
	deploymentLister extensionslisters.DeploymentLister,
	serviceAccountLister corelisters.ServiceAccountLister,
	serviceLister corelisters.ServiceLister,
	nodePoolController nodepool.Interface,
	serviceAccountControl ElasticsearchClusterServiceAccountControl,
	clientServiceControl ElasticsearchClusterServiceControl,
	discoveryServiceControl ElasticsearchClusterServiceControl,
	recorder record.EventRecorder,
) ElasticsearchClusterControl {
	return &defaultElasticsearchClusterControl{
		statefulSetLister:       statefulSetLister,
		deploymentLister:        deploymentLister,
		serviceAccountLister:    serviceAccountLister,
		serviceLister:           serviceLister,
		nodePoolController:      nodePoolController,
		serviceAccountControl:   serviceAccountControl,
		clientServiceControl:    clientServiceControl,
		discoveryServiceControl: discoveryServiceControl,
		recorder:                recorder,
	}
}

func (e *defaultElasticsearchClusterControl) SyncElasticsearchCluster(
	c v1alpha1.ElasticsearchCluster,
) error {
	var err error

	if err = e.syncServiceAccount(c); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	if err = e.syncService(c, e.clientServiceControl); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	if err = e.syncService(c, e.discoveryServiceControl); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	// TODO: handle status
	if _, err = e.nodePoolController.Sync(c); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	e.recordClusterEvent("sync", c, err)
	return nil
}

func (e *defaultElasticsearchClusterControl) syncService(c v1alpha1.ElasticsearchCluster, ctrl ElasticsearchClusterServiceControl) error {
	exists, needsUpdate, err := e.serviceNeedsUpdate(c, ctrl.NameSuffix())

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	// TODO: switch to proper deletion handling logic using tombstones
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

func (e *defaultElasticsearchClusterControl) syncServiceAccount(c v1alpha1.ElasticsearchCluster) error {
	exists, needsUpdate, err := e.serviceAccountNeedsUpdate(c)

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
		err = e.serviceAccountControl.DeleteElasticsearchClusterServiceAccount(c)
		break
	case exists:
		err = e.serviceAccountControl.UpdateElasticsearchClusterServiceAccount(c)
		break
	default:
		err = e.serviceAccountControl.CreateElasticsearchClusterServiceAccount(c)
	}

	if err != nil {
		if errors.IsTransient(err) {
			return err
		}

		return nil
	}

	return nil
}

func (e *defaultElasticsearchClusterControl) serviceNeedsUpdate(c v1alpha1.ElasticsearchCluster, nameSuffix string) (exists, needsUpdate bool, err error) {
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

func (e *defaultElasticsearchClusterControl) serviceAccountNeedsUpdate(c v1alpha1.ElasticsearchCluster) (exists, needsUpdate bool, err error) {
	svcAcctName := resourceBaseName(c)
	svcAccts, err := e.serviceAccountLister.ServiceAccounts(c.Namespace).List(labels.Everything())

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, false, errors.Transient(fmt.Errorf("error getting list of service accounts: %s", err.Error()))
		}

		return false, false, errors.Transient(fmt.Errorf("error getting serviceaccounts from apiserver: %s", err.Error()))
	}

	if len(svcAccts) == 0 {
		return false, true, nil
	}

	for _, svcAcct := range svcAccts {
		// TODO: switch this to use UIDs set as annotations on the ElasticsearchCluster?
		if svcAcct.Name == svcAcctName {
			if !isManagedByCluster(c, svcAcct.ObjectMeta) {
				return false, false, fmt.Errorf("service account '%s' found but not managed by cluster", svcAcctName)
			}
			return true, false, nil
		}
	}

	return false, true, nil
}

var notFoundErr = fmt.Errorf("resource not found")

func (e *defaultElasticsearchClusterControl) deploymentForNodePool(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (*extensions.Deployment, error) {
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

// recordClusterEvent records an event for verb applied to the ElasticsearchCluster. If err is nil the generated event will
// have a reason of apiv1.EventTypeNormal. If err is not nil the generated event will have a reason of apiv1.EventTypeWarning.
func (e *defaultElasticsearchClusterControl) recordClusterEvent(verb string, cluster v1alpha1.ElasticsearchCluster, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in ElasticsearchCluster %s successful",
			strings.ToLower(verb), cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
