package serviceaccount

import (
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type control struct {
	kubeClient           kubernetes.Interface
	serviceAccountLister corelisters.ServiceAccountLister
	recorder             record.EventRecorder
}

var _ Interface = &control{}

func NewControl(
	kubeClient kubernetes.Interface,
	serviceAccountLister corelisters.ServiceAccountLister,
	recorder record.EventRecorder,
) *control {
	return &control{
		kubeClient:           kubeClient,
		serviceAccountLister: serviceAccountLister,
		recorder:             recorder,
	}
}

func (c *control) Sync(cluster *v1alpha1.CassandraCluster) error {
	newAccount := ServiceAccountForCluster(cluster)
	client := c.kubeClient.CoreV1().ServiceAccounts(newAccount.Namespace)
	existingAccount, err := c.serviceAccountLister.
		ServiceAccounts(newAccount.Namespace).
		Get(newAccount.Name)
	if err == nil {
		return util.OwnerCheck(existingAccount, cluster)
	}
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	_, err = client.Create(newAccount)
	return err
}

func ServiceAccountForCluster(cluster *v1alpha1.CassandraCluster) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ServiceAccountName(cluster),
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				util.NewControllerRef(cluster),
			},
			Labels: util.ClusterLabels(cluster),
		},
	}
}
