package v1

import (
	"github.com/jetstack-experimental/navigator/pkg/informers/internalinterfaces"
)

type Interface interface {
	ElasticsearchCluster() ElasticsearchClusterInformer
}

var _ Interface = &version{}

type version struct {
	internalinterfaces.SharedInformerFactory
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory) Interface {
	return &version{f}
}

// StatefulSets returns a StatefulSetInformer.
func (v *version) ElasticsearchCluster() ElasticsearchClusterInformer {
	return &elasticsearchClusterInformer{factory: v.SharedInformerFactory}
}
