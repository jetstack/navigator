package elasticsearch

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/record"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
)

type ElasticsearchClusterServiceControl interface {
	CreateElasticsearchClusterService(*v1.ElasticsearchCluster) error
	UpdateElasticsearchClusterService(*v1.ElasticsearchCluster) error
	DeleteElasticsearchClusterService(*v1.ElasticsearchCluster) error
	NameSuffix() string
}

type defaultElasticsearchClusterServiceControl struct {
	kubeClient *kubernetes.Clientset

	nameSuffix string
	enableHTTP bool
	roles      []string

	recorder record.EventRecorder
}

var _ ElasticsearchClusterServiceControl = &defaultElasticsearchClusterServiceControl{}

func NewElasticsearchClusterServiceControl(
	kubeClient *kubernetes.Clientset,
	recorder record.EventRecorder,
	nameSuffix string,
	enableHTTP bool,
	roles ...string,
) ElasticsearchClusterServiceControl {
	return &defaultElasticsearchClusterServiceControl{
		kubeClient: kubeClient,
		enableHTTP: enableHTTP,
		nameSuffix: nameSuffix,
		roles:      roles,
		recorder:   recorder,
	}
}

func (e *defaultElasticsearchClusterServiceControl) NameSuffix() string {
	return e.nameSuffix
}

func (e *defaultElasticsearchClusterServiceControl) CreateElasticsearchClusterService(c *v1.ElasticsearchCluster) (err error) {
	svc := clusterService(c, e.NameSuffix(), e.enableHTTP, e.roles...)

	svc, err = e.kubeClient.Core().Services(c.Namespace).Create(svc)

	if err != nil {
		e.recordEvent("create", c, svc, err)
		return fmt.Errorf("error creating service: %s", err.Error())
	}

	e.recordEvent("create", c, svc, err)
	return nil
}

func (e *defaultElasticsearchClusterServiceControl) UpdateElasticsearchClusterService(c *v1.ElasticsearchCluster) (err error) {
	svc := clusterService(c, e.NameSuffix(), e.enableHTTP, e.roles...)

	svc, err = e.kubeClient.Core().Services(c.Namespace).Update(svc)

	if err != nil {
		e.recordEvent("update", c, svc, err)
		return fmt.Errorf("error creating service: %s", err.Error())
	}

	e.recordEvent("update", c, svc, err)
	return nil
}

func (e *defaultElasticsearchClusterServiceControl) DeleteElasticsearchClusterService(c *v1.ElasticsearchCluster) error {
	svc := clusterService(c, e.NameSuffix(), e.enableHTTP, e.roles...)

	err := e.kubeClient.Core().Services(c.Namespace).Delete(svc.Name, &metav1.DeleteOptions{})

	if err != nil {
		e.recordEvent("delete", c, svc, err)
		return fmt.Errorf("error deleting service: %s", err.Error())
	}

	e.recordEvent("delete", c, svc, err)
	return nil
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterServiceControl) recordEvent(verb string, cluster *v1.ElasticsearchCluster, svc *apiv1.Service, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Service %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), svc.Name, cluster.Name)
		e.recorder.Event(cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Service %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), svc.Name, cluster.Name, err)
		e.recorder.Event(cluster, apiv1.EventTypeWarning, reason, message)
	}
}
