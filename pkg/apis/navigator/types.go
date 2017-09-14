package navigator

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// In this file we define the outer containing types for the ElasticsearchCluster
// type. We could import these directly into message types defined in the types.proto
// file, but this is still TODO

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchClusterList defines a List type for our custom ElasticsearchCluster type.
// This is needed in order to make List operations work.
type ElasticsearchClusterList struct {
	// we embed these types so that ElasticsearchClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta
	metav1.ListMeta

	Items []ElasticsearchCluster
}

type ElasticsearchClusterSpec struct {
	Plugins   []ElasticsearchClusterPlugin
	NodePools []ElasticsearchClusterNodePool
	Pilot     ElasticsearchPilotImage
	Image     ElasticsearchImage
	Sysctl    []string
}

type ElasticsearchClusterPlugin struct {
	Name string
}

type ElasticsearchClusterNodePool struct {
	Name         string
	Replicas     int32
	Roles        []string
	NodeSelector map[string]string
	Resources    *v1.ResourceRequirements
	Persistence  ElasticsearchClusterPersistenceConfig
}

type ElasticsearchClusterPersistenceConfig struct {
	Enabled      bool
	Size         string
	StorageClass string
}

type ImageSpec struct {
	Repository string
	Tag        string
	PullPolicy string
}

type ElasticsearchPilotImage struct {
	ImageSpec
}

type ElasticsearchImage struct {
	ImageSpec
	FsGroup int64
}
