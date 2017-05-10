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

type ElasticsearchClusterServiceControl interface {
	CreateElasticsearchClusterService(v1alpha1.ElasticsearchCluster) error
	UpdateElasticsearchClusterService(v1alpha1.ElasticsearchCluster) error
	DeleteElasticsearchClusterService(v1alpha1.ElasticsearchCluster) error
	NameSuffix() string
}

type ServiceControlConfig struct {
	ClusterIP   string
	NameSuffix  string
	EnableHTTP  bool
	Annotations map[string]string
	Roles       []string
}

type defaultElasticsearchClusterServiceControl struct {
	kubeClient *kubernetes.Clientset

	config ServiceControlConfig

	recorder record.EventRecorder
}

var _ ElasticsearchClusterServiceControl = &defaultElasticsearchClusterServiceControl{}

func NewElasticsearchClusterServiceControl(
	kubeClient *kubernetes.Clientset,
	recorder record.EventRecorder,
	config ServiceControlConfig,
) ElasticsearchClusterServiceControl {
	return &defaultElasticsearchClusterServiceControl{
		kubeClient: kubeClient,
		config:     config,
		recorder:   recorder,
	}
}

func (e *defaultElasticsearchClusterServiceControl) NameSuffix() string {
	return e.config.NameSuffix
}

func (e *defaultElasticsearchClusterServiceControl) CreateElasticsearchClusterService(c v1alpha1.ElasticsearchCluster) (err error) {
	svc := clusterService(c, e.NameSuffix(), e.config.EnableHTTP, e.config.Annotations, e.config.Roles...)

	svc, err = e.kubeClient.Core().Services(c.Namespace).Create(svc)

	if err != nil {
		e.recordEvent("create", c, svc, err)
		return fmt.Errorf("error creating service: %s", err.Error())
	}

	e.recordEvent("create", c, svc, err)
	return nil
}

func (e *defaultElasticsearchClusterServiceControl) UpdateElasticsearchClusterService(c v1alpha1.ElasticsearchCluster) (err error) {
	svc := clusterService(c, e.NameSuffix(), e.config.EnableHTTP, e.config.Annotations, e.config.Roles...)

	svc, err = e.kubeClient.Core().Services(c.Namespace).Update(svc)

	if err != nil {
		e.recordEvent("update", c, svc, err)
		return fmt.Errorf("error creating service: %s", err.Error())
	}

	e.recordEvent("update", c, svc, err)
	return nil
}

func (e *defaultElasticsearchClusterServiceControl) DeleteElasticsearchClusterService(c v1alpha1.ElasticsearchCluster) error {
	svc := clusterService(c, e.NameSuffix(), e.config.EnableHTTP, e.config.Annotations, e.config.Roles...)

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
