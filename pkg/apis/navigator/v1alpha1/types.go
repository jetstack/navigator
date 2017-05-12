package v1alpha1

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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ElasticsearchClusterSpec   `json:"spec"`
	Status ElasticsearchClusterStatus `json:"status"`
}

type ElasticsearchClusterStatus struct {
}

// ElasticsearchClusterList defines a List type for our custom ElasticsearchCluster type.
// This is needed in order to make List operations work.
type ElasticsearchClusterList struct {
	// we embed these types so that ElasticsearchClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ElasticsearchCluster `json:"items"`
}

// ElasticsearchClusterSpec describes a specification for an ElasticsearchCluster
type ElasticsearchClusterSpec struct {
	Version   string                         `json:"version"`
	Plugins   []ElasticsearchClusterPlugin   `json:"plugins"`
	NodePools []ElasticsearchClusterNodePool `json:"nodePools"`
	// TODO: Remove this field
	Image  ElasticsearchImage `json:"image"`
	Sysctl []string           `json:"sysctl"`
}

// ElasticsearchClusterPlugin describes a specification of an ElasticsearchCluster plugin
// You must ensure the plugin is compatible with the version of Elasticsearch being deployed
// else the cluster will not deploy successfully
type ElasticsearchClusterPlugin struct {
	Name string `json:"name"`
}

// ElasticsearchClusterNodePool describes a node pool within an ElasticsearchCluster.
// The nodes in this pool will be configured to be of the specified roles
type ElasticsearchClusterNodePool struct {
	Name        string                                 `json:"name"`
	Replicas    int32                                  `json:"replicas"`
	Roles       []string                               `json:"roles"`
	Resources   *v1.ResourceRequirements               `json:"resources,omitempty"`
	Persistence *ElasticsearchClusterPersistenceConfig `json:"persistence,omitempty"`
}

type ElasticsearchClusterPersistenceConfig struct {
	Size         string `json:"size"`
	StorageClass string `json:"storageClass"`
}

// TODO: Remove this struct
type ElasticsearchImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
	FsGroup    int64  `json:"fsGroup"`
}
