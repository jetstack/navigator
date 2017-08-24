package couchbase

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

type CouchbaseClusterControl interface {
	SyncCouchbaseCluster(v1alpha1.CouchbaseCluster) error
}

type CouchbaseClusterNodePoolControl interface {
	CreateElasticsearchClusterNodePool(v1alpha1.ElasticsearchCluster, v1alpha1.CouchbaseClusterNodePool) error
	UpdateElasticsearchClusterNodePool(v1alpha1.ElasticsearchCluster, v1alpha1.CouchbaseClusterNodePool) error
	DeleteElasticsearchClusterNodePool(v1alpha1.ElasticsearchCluster, v1alpha1.CouchbaseClusterNodePool) error
}

type CouchbaseClusterServiceAccountControl interface {
	CreateElasticsearchClusterServiceAccount(v1alpha1.CouchbaseCluster) error
	UpdateElasticsearchClusterServiceAccount(v1alpha1.CouchbaseCluster) error
	DeleteElasticsearchClusterServiceAccount(v1alpha1.CouchbaseCluster) error
}

type CouchbaseClusterServiceControl interface {
	CreateElasticsearchClusterService(v1alpha1.CouchbaseCluster) error
	UpdateElasticsearchClusterService(v1alpha1.CouchbaseCluster) error
	DeleteElasticsearchClusterService(v1alpha1.CouchbaseCluster) error
	NameSuffix() string
}
