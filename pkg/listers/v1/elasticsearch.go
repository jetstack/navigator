package v1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
)

type ElasticsearchClusterLister interface {
	List(selector labels.Selector) (ret []*v1.ElasticsearchCluster, err error)
	ElasticsearchClusters(namespace string) ElasticsearchClusterNamespaceLister
}

var _ ElasticsearchClusterLister = &elasticsearchClusterLister{}

type elasticsearchClusterLister struct {
	indexer cache.Indexer
}

func NewElasticsearchClusterLister(indexer cache.Indexer) ElasticsearchClusterLister {
	return &elasticsearchClusterLister{indexer: indexer}
}

func (s *elasticsearchClusterLister) List(selector labels.Selector) (ret []*v1.ElasticsearchCluster, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ElasticsearchCluster))
	})
	return ret, err
}

func (s *elasticsearchClusterLister) ElasticsearchClusters(namespace string) ElasticsearchClusterNamespaceLister {
	return &elasticsearchClusterNamespaceLister{indexer: s.indexer, namespace: namespace}
}

type ElasticsearchClusterNamespaceLister interface {
	List(selector labels.Selector) (ret []*v1.ElasticsearchCluster, err error)
	Get(name string) (*v1.ElasticsearchCluster, error)
}

var _ ElasticsearchClusterNamespaceLister = &elasticsearchClusterNamespaceLister{}

type elasticsearchClusterNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

func (s *elasticsearchClusterNamespaceLister) List(selector labels.Selector) (ret []*v1.ElasticsearchCluster, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ElasticsearchCluster))
	})
	return ret, err
}

func (s *elasticsearchClusterNamespaceLister) Get(name string) (*v1.ElasticsearchCluster, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("elasticsearchcluster"), name)
	}
	return obj.(*v1.ElasticsearchCluster), nil
}
