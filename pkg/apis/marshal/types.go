package marshal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// In this file we define the outer containing types for the ElasticsearchCluster
// type. We could import these directly into message types defined in the types.proto
// file, but this is still TODO

// +genclient=true

// ElasticsearchCluster describes a specification for an Elasticsearch cluster
type ElasticsearchCluster struct {
	// we embed these types so the ElasticsearchCluster implements runtime.Object
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ElasticsearchClusterSpec
	Status ElasticsearchClusterStatus
}

type ElasticsearchClusterStatus struct {
}

// ElasticsearchClusterList defines a List type for our custom ElasticsearchCluster type.
// This is needed in order to make List operations work.
type ElasticsearchClusterList struct {
	// we embed these types so that ElasticsearchClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta
	metav1.ListMeta

	Items []ElasticsearchCluster
}

type ElasticsearchClusterSpec struct {
	Version   string
	Plugins   []ElasticsearchClusterPlugin
	NodePools []ElasticsearchClusterNodePool
	Image     ElasticsearchImage
	Sysctl    []string
}

type ElasticsearchClusterPlugin struct {
	Name string
}

type ElasticsearchClusterNodePool struct {
	Name        string
	Replicas    int32
	Roles       []string
	Resources   *v1.ResourceRequirements
	Persistence *ElasticsearchClusterPersistenceConfig
}

type ElasticsearchClusterPersistenceConfig struct {
	Size         string
	StorageClass string
}

type ElasticsearchImage struct {
	Repository string
	Tag        string
	PullPolicy string
	FsGroup    int64
}
