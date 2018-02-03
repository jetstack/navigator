package serviceaccount

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

type defaultElasticsearchClusterServiceAccountControl struct {
	kubeClient           kubernetes.Interface
	serviceAccountLister listersv1.ServiceAccountLister

	recorder record.EventRecorder
}

type Interface interface {
	Sync(*v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error)
}

var _ Interface = &defaultElasticsearchClusterServiceAccountControl{}

func NewController(
	kubeClient kubernetes.Interface,
	serviceAccountLister listersv1.ServiceAccountLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultElasticsearchClusterServiceAccountControl{
		kubeClient:           kubeClient,
		serviceAccountLister: serviceAccountLister,
		recorder:             recorder,
	}
}

func (e *defaultElasticsearchClusterServiceAccountControl) Sync(c *v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error) {
	// lookup existing ServiceAccount with appropriate labels for np in cluster c
	// if multiple exist, exit with an error
	// if one exists, return
	// if none exist, create one
	sel, err := util.SelectorForCluster(c.Name)
	if err != nil {
		return c.Status, fmt.Errorf("error creating label selector for cluster '%s': %s", c.Name, err.Error())
	}
	svcAccts, err := e.serviceAccountLister.ServiceAccounts(c.Namespace).List(sel)
	if err != nil {
		return c.Status, err
	}
	// if more than one serviceaccount matches the labels, exit here to be safe
	if len(svcAccts) > 1 {
		return c.Status, fmt.Errorf("multiple service accounts match label selector (%s)", sel.String())
	}
	expected := clusterServiceAccount(c)
	if len(svcAccts) == 0 {
		_, err := e.kubeClient.CoreV1().ServiceAccounts(c.Namespace).Create(expected)
		return c.Status, err
	}
	return c.Status, nil
}
