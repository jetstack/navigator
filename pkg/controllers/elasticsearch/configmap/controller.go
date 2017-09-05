package configmap

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

type defaultElasticsearchClusterConfigMapControl struct {
	kubeClient      kubernetes.Interface
	configMapLister corelisters.ConfigMapLister

	recorder record.EventRecorder
}

var _ Interface = &defaultElasticsearchClusterConfigMapControl{}

func NewController(
	kubeClient kubernetes.Interface,
	configMapLister corelisters.ConfigMapLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultElasticsearchClusterConfigMapControl{
		kubeClient:      kubeClient,
		configMapLister: configMapLister,
		recorder:        recorder,
	}
}

func (e *defaultElasticsearchClusterConfigMapControl) Sync(c *v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error) {
	if err := e.ensureConfigMap(esConfigConfigMap(c)); err != nil {
		return c.Status, fmt.Errorf("error ensuring config ConfigMap: %s", err.Error())
	}
	return c.Status, nil
}

func (e *defaultElasticsearchClusterConfigMapControl) ensureConfigMap(configMap *apiv1.ConfigMap) error {
	_, err := e.configMapLister.ConfigMaps(configMap.Namespace).Get(configMap.Name)
	if k8sErrors.IsNotFound(err) {
		_, err := e.kubeClient.CoreV1().ConfigMaps(configMap.Namespace).Create(configMap)
		return err
	}
	return err
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterConfigMapControl) recordEvent(verb string, cluster v1alpha1.ElasticsearchCluster, cm *apiv1.ConfigMap, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s ConfigMap %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), cm.Name, cluster.Name)
		e.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s ConfigMap %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), cm.Name, cluster.Name, err)
		e.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
