package util

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	corelisters "k8s.io/client-go/listers/core/v1"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

func SetStandardServiceAttributes(
	cluster *v1alpha1.CassandraCluster,
	service *apiv1.Service,
) *apiv1.Service {
	service.SetNamespace(cluster.Namespace)
	service.SetLabels(util.ClusterLabels(cluster))
	service.SetOwnerReferences([]metav1.OwnerReference{
		util.NewControllerRef(cluster),
	})
	service.Spec.Selector = util.NodePoolLabels(cluster, "")
	return service
}

type ServiceCreator func(cluster *v1alpha1.CassandraCluster) *apiv1.Service
type ServiceUpdater func(cluster *v1alpha1.CassandraCluster, service *apiv1.Service) *apiv1.Service

func SyncService(
	cluster *v1alpha1.CassandraCluster,
	kubeClient kubernetes.Interface,
	serviceLister corelisters.ServiceLister,
	createService ServiceCreator,
	updateService ServiceUpdater,
) error {
	svc := createService(cluster)
	client := kubeClient.CoreV1().Services(svc.Namespace)
	existingSvc, err := serviceLister.Services(svc.Namespace).Get(svc.Name)
	if k8sErrors.IsNotFound(err) {
		_, err = client.Create(svc)
		return err
	}
	if err != nil {
		return err
	}
	err = util.OwnerCheck(existingSvc, cluster)
	if err != nil {
		return err
	}
	updatedService := updateService(cluster, existingSvc)
	_, err = client.Update(updatedService)
	return err
}
