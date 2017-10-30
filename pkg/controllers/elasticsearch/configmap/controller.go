package configmap

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
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
	for _, np := range c.Spec.NodePools {
		if err := e.ensureConfigMap(esConfigConfigMap(c, &np)); err != nil {
			return c.Status, fmt.Errorf("error ensuring %q/%q nodepool configmap: %s", c.Name, np.Name, err.Error())
		}
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
