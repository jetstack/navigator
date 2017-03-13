package v1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
	"gitlab.jetstack.net/marshal/colonel/pkg/informers/internalinterfaces"
	listersv1 "gitlab.jetstack.net/marshal/colonel/pkg/listers/v1"
)

type ElasticsearchClusterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() listersv1.ElasticsearchClusterLister
}

var _ ElasticsearchClusterInformer = &elasticsearchClusterInformer{}

type elasticsearchClusterInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

func newElasticsearchClusterInformer(cl *rest.RESTClient, resyncPeriod time.Duration) cache.SharedIndexInformer {
	sharedIndexInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (obj runtime.Object, err error) {
				return cl.
					Get().
					Resource("elasticsearchclusters").
					Namespace(api.NamespaceAll).
					// VersionedParams(&options, api.ParameterCodec).
					Do().Get()
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return cl.
					Get().
					Prefix("watch").
					Resource("elasticsearchclusters").
					Namespace(api.NamespaceAll).
					// VersionedParams(&options, api.ParameterCodec).
					Watch()
			},
		},
		&v1.ElasticsearchCluster{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	return sharedIndexInformer
}

func (f *elasticsearchClusterInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&v1.ElasticsearchCluster{}, newElasticsearchClusterInformer)
}

func (f *elasticsearchClusterInformer) Lister() listersv1.ElasticsearchClusterLister {
	return listersv1.NewElasticsearchClusterLister(f.Informer().GetIndexer())
}
