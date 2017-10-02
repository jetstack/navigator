package elasticsearch

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/configmap"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/nodepool"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/service"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/serviceaccount"
)

const (
	errorSync = "ErrSync"

	successSync = "SuccessSync"

	messageErrorSyncServiceAccount = "Error syncing service account: %s"
	messageErrorSyncConfigMap      = "Error syncing config map: %s"
	messageErrorSyncService        = "Error syncing service: %s"
	messageErrorSyncNodePools      = "Error syncing node pools: %s"
	messageSuccessSync             = "Successfully synced ElasticsearchCluster"
)

type ControlInterface interface {
	Sync(*v1alpha1.ElasticsearchCluster) error
}

var _ ControlInterface = &defaultElasticsearchClusterControl{}

type defaultElasticsearchClusterControl struct {
	kubeClient *kubernetes.Clientset

	statefulSetLister    appslisters.StatefulSetLister
	serviceAccountLister corelisters.ServiceAccountLister
	serviceLister        corelisters.ServiceLister

	nodePoolControl       nodepool.Interface
	configMapControl      configmap.Interface
	serviceAccountControl serviceaccount.Interface
	serviceControl        service.Interface

	recorder record.EventRecorder
}

var _ ControlInterface = &defaultElasticsearchClusterControl{}

func NewController(
	statefulSetLister appslisters.StatefulSetLister,
	serviceAccountLister corelisters.ServiceAccountLister,
	serviceLister corelisters.ServiceLister,
	nodePoolControl nodepool.Interface,
	configMapControl configmap.Interface,
	serviceAccountControl serviceaccount.Interface,
	serviceControl service.Interface,
	recorder record.EventRecorder,
) ControlInterface {
	return &defaultElasticsearchClusterControl{
		statefulSetLister:     statefulSetLister,
		serviceAccountLister:  serviceAccountLister,
		serviceLister:         serviceLister,
		nodePoolControl:       nodePoolControl,
		configMapControl:      configMapControl,
		serviceAccountControl: serviceAccountControl,
		serviceControl:        serviceControl,
		recorder:              recorder,
	}
}

func (e *defaultElasticsearchClusterControl) Sync(c *v1alpha1.ElasticsearchCluster) error {
	var err error

	// TODO: handle status
	if _, err = e.serviceAccountControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncServiceAccount, err.Error())
		return err
	}

	// TODO: handle status
	if _, err = e.configMapControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncConfigMap, err.Error())
		return err
	}

	// TODO: handle status
	if _, err = e.serviceControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncService, err.Error())
		return err
	}

	// TODO: handle status
	if _, err = e.nodePoolControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncNodePools, err.Error())
		return err
	}

	e.recorder.Event(c, apiv1.EventTypeNormal, successSync, messageSuccessSync)

	return nil
}
