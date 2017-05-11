package elasticsearch

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/marshal/v1alpha1"
)

type ElasticsearchClusterServiceAccountControl interface {
	CreateElasticsearchClusterServiceAccount(v1alpha1.ElasticsearchCluster) error
	UpdateElasticsearchClusterServiceAccount(v1alpha1.ElasticsearchCluster) error
	DeleteElasticsearchClusterServiceAccount(v1alpha1.ElasticsearchCluster) error
}

type defaultElasticsearchClusterServiceAccountControl struct {
	kubeClient *kubernetes.Clientset

	recorder record.EventRecorder
}

var _ ElasticsearchClusterServiceAccountControl = &defaultElasticsearchClusterServiceAccountControl{}

func NewElasticsearchClusterServiceAccountControl(
	kubeClient *kubernetes.Clientset,
	recorder record.EventRecorder,
) ElasticsearchClusterServiceAccountControl {
	return &defaultElasticsearchClusterServiceAccountControl{
		kubeClient: kubeClient,
		recorder:   recorder,
	}
}

func (e *defaultElasticsearchClusterServiceAccountControl) CreateElasticsearchClusterServiceAccount(c v1alpha1.ElasticsearchCluster) (err error) {
	svcAcct := clusterServiceAccount(c)

	svcAcct, err = e.kubeClient.Core().ServiceAccounts(c.Namespace).Create(svcAcct)

	if err != nil {
		e.recordEvent("create", c, svcAcct, err)
		return fmt.Errorf("error creating serviceaccount: %s", err.Error())
	}

	e.recordEvent("create", c, svcAcct, err)
	return nil
}

func (e *defaultElasticsearchClusterServiceAccountControl) UpdateElasticsearchClusterServiceAccount(c v1alpha1.ElasticsearchCluster) (err error) {
	svcAcct := clusterServiceAccount(c)

	svcAcct, err = e.kubeClient.Core().ServiceAccounts(c.Namespace).Update(svcAcct)

	if err != nil {
		e.recordEvent("update", c, svcAcct, err)
		return fmt.Errorf("error creating serviceaccount: %s", err.Error())
	}

	e.recordEvent("update", c, svcAcct, err)
	return nil
}

func (e *defaultElasticsearchClusterServiceAccountControl) DeleteElasticsearchClusterServiceAccount(c v1alpha1.ElasticsearchCluster) error {
	svcAcct := clusterServiceAccount(c)

	err := e.kubeClient.Core().ServiceAccounts(c.Namespace).Delete(svcAcct.Name, &metav1.DeleteOptions{})

	if err != nil {
		e.recordEvent("delete", c, svcAcct, err)
		return fmt.Errorf("error deleting serviceaccount: %s", err.Error())
	}

	e.recordEvent("delete", c, svcAcct, err)
	return nil
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
