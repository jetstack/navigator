package service

import (
	"fmt"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"

	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	SeedLabelKey   = "navigator.jetstack.io/cassandra-seed"
	SeedLabelValue = "true"

	// This field is deprecated. v1.Service.PublishNotReadyAddresses will replace it.
	// See https://github.com/kubernetes/kubernetes/blob/v1.10.0/pkg/controller/endpoint/endpoints_controller.go#L68
	tolerateUnreadyEndpointsAnnotationKey = "service.alpha.kubernetes.io/tolerate-unready-endpoints"
)

type serviceFactory func(*v1alpha1.CassandraCluster) *apiv1.Service

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type control struct {
	kubeClient     kubernetes.Interface
	serviceLister  corelisters.ServiceLister
	recorder       record.EventRecorder
	serviceFactory serviceFactory
}

var _ Interface = &control{}

func NewControl(
	kubeClient kubernetes.Interface,
	serviceLister corelisters.ServiceLister,
	recorder record.EventRecorder,
	serviceFactory serviceFactory,
) Interface {
	return &control{
		kubeClient:     kubeClient,
		serviceLister:  serviceLister,
		recorder:       recorder,
		serviceFactory: serviceFactory,
	}
}

func (c *control) Sync(cluster *v1alpha1.CassandraCluster) error {
	service := c.serviceFactory(cluster)
	existingService, err := c.serviceLister.Services(service.Namespace).Get(service.Name)
	if err == nil {
		return util.OwnerCheck(existingService, cluster)
	}
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	_, err = c.kubeClient.CoreV1().Services(service.Namespace).Create(service)
	return err
}

func NodesServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
) *apiv1.Service {
	labels := util.ClusterLabels(cluster)
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-nodes", util.ResourceBaseName(cluster)),
			Namespace:       cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
			Labels:          labels,
		},
		Spec: apiv1.ServiceSpec{
			ClusterIP: "None",
			Type:      apiv1.ServiceTypeClusterIP,
			Selector:  labels,
			// Headless service should not require a port.
			// But without it, DNS records are not registered.
			// See https://github.com/kubernetes/kubernetes/issues/55158
			Ports: []apiv1.ServicePort{{Port: 65535}},
		},
	}
}

var _ serviceFactory = NodesServiceForCluster

func SeedsServiceForCluster(cluster *v1alpha1.CassandraCluster) *apiv1.Service {
	labels := util.ClusterLabels(cluster)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.SeedsServiceName(cluster),
			Namespace: cluster.Namespace,
			Annotations: map[string]string{
				tolerateUnreadyEndpointsAnnotationKey: "true",
			},
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
			Labels:          labels,
		},
		Spec: apiv1.ServiceSpec{
			ClusterIP: "None",
			Type:      apiv1.ServiceTypeClusterIP,
			Selector:  labels,
			// Headless service should not require a port.
			// But without it, DNS records are not registered.
			// See https://github.com/kubernetes/kubernetes/issues/55158
			Ports: []apiv1.ServicePort{{Port: 65535}},
			// This ensures that seed node addresses are published immediately.
			// And are available during C* node start up.
			// Even for seed nodes them selves.
			PublishNotReadyAddresses: true,
		},
	}
	// Only mark nodes explicitly labeled as seeds as seed nodes
	service.Spec.Selector[SeedLabelKey] = SeedLabelValue
	return service
}

var _ serviceFactory = SeedsServiceForCluster
