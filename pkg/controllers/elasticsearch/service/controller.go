package service

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

type Interface interface {
	Sync(*v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error)
}

type defaultElasticsearchClusterServiceControl struct {
	kubeClient kubernetes.Interface
	svcLister  corelisters.ServiceLister

	recorder record.EventRecorder
}

var _ Interface = &defaultElasticsearchClusterServiceControl{}

func NewController(
	kubeClient kubernetes.Interface,
	svcLister corelisters.ServiceLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultElasticsearchClusterServiceControl{
		kubeClient: kubeClient,
		svcLister:  svcLister,
		recorder:   recorder,
	}
}

func (e *defaultElasticsearchClusterServiceControl) Sync(c *v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error) {
	if err := e.ensureService(discoveryService(c)); err != nil {
		return c.Status, fmt.Errorf("error ensuring discovery service: %s", err.Error())
	}
	if err := e.ensureService(clientService(c)); err != nil {
		return c.Status, fmt.Errorf("error ensuring client service: %s", err.Error())
	}
	return c.Status, nil
}

func (e *defaultElasticsearchClusterServiceControl) ensureService(svc *apiv1.Service) error {
	_, err := e.svcLister.Services(svc.Namespace).Get(svc.Name)
	if k8sErrors.IsNotFound(err) {
		_, err := e.kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
		return err
	}
	return err
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterServiceControl) recordEvent(verb string, cluster v1alpha1.ElasticsearchCluster, svc *apiv1.Service, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Service %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), svc.Name, cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Service %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), svc.Name, cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
