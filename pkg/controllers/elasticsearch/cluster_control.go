package elasticsearch

import (
	"fmt"
	"strings"

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
		e.recordClusterEvent("sync", c, err)
		return err
	}

	// TODO: handle status
	if _, err = e.configMapControl.Sync(c); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	// TODO: handle status
	if _, err = e.serviceControl.Sync(c); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	// TODO: handle status
	if _, err = e.nodePoolControl.Sync(c); err != nil {
		e.recordClusterEvent("sync", c, err)
		return err
	}

	e.recordClusterEvent("sync", c, err)
	return nil
}

// recordClusterEvent records an event for verb applied to the ElasticsearchCluster. If err is nil the generated event will
// have a reason of apiv1.EventTypeNormal. If err is not nil the generated event will have a reason of apiv1.EventTypeWarning.
func (e *defaultElasticsearchClusterControl) recordClusterEvent(verb string, cluster *v1alpha1.ElasticsearchCluster, err error) {
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
