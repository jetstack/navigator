package serviceaccount

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
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
	sel, err := util.SelectorForCluster(c)
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

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterServiceAccountControl) recordEvent(verb string, cluster v1alpha1.ElasticsearchCluster, svcAcct *apiv1.ServiceAccount, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s ServiceAccount %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), svcAcct.Name, cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s ServiceAccount %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), svcAcct.Name, cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
