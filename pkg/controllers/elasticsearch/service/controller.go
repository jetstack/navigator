package service

import (
	"fmt"

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
