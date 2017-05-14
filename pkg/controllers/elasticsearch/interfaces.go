package elasticsearch

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

type ElasticsearchClusterControl interface {
	SyncElasticsearchCluster(v1alpha1.ElasticsearchCluster) error
}

type ElasticsearchClusterNodePoolControl interface {
	CreateElasticsearchClusterNodePool(v1alpha1.ElasticsearchCluster, v1alpha1.ElasticsearchClusterNodePool) error
	UpdateElasticsearchClusterNodePool(v1alpha1.ElasticsearchCluster, v1alpha1.ElasticsearchClusterNodePool) error
	DeleteElasticsearchClusterNodePool(v1alpha1.ElasticsearchCluster, v1alpha1.ElasticsearchClusterNodePool) error
}

type ElasticsearchClusterServiceAccountControl interface {
	CreateElasticsearchClusterServiceAccount(v1alpha1.ElasticsearchCluster) error
	UpdateElasticsearchClusterServiceAccount(v1alpha1.ElasticsearchCluster) error
	DeleteElasticsearchClusterServiceAccount(v1alpha1.ElasticsearchCluster) error
}

type ElasticsearchClusterServiceControl interface {
	CreateElasticsearchClusterService(v1alpha1.ElasticsearchCluster) error
	UpdateElasticsearchClusterService(v1alpha1.ElasticsearchCluster) error
	DeleteElasticsearchClusterService(v1alpha1.ElasticsearchCluster) error
	NameSuffix() string
}
