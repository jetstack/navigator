package elasticsearch

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

type ElasticsearchClusterControl interface {
	SyncElasticsearchCluster(*v1alpha1.ElasticsearchCluster) error
}
