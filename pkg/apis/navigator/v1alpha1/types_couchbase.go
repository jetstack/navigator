package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// In this file we define the outer containing types for the CouchbaseCluster
// type. We could import these directly into message types defined in the types.proto
// file, but this is still TODO

// +genclient=true

// CouchbaseCluster describes a specification for an Couchbase cluster
type CouchbaseCluster struct {
	// we embed these types so the CouchbaseCluster implements runtime.Object
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CouchbaseClusterSpec   `json:"spec"`
	Status CouchbaseClusterStatus `json:"status"`
}

type CouchbaseClusterStatus struct {
}

// CouchbaseClusterList defines a List type for our custom CouchbaseCluster type.
// This is needed in order to make List operations work.
type CouchbaseClusterList struct {
	// we embed these types so that CouchbaseClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CouchbaseCluster `json:"items"`
}

// CouchbaseClusterSpec describes a specification for an CouchbaseCluster
type CouchbaseClusterSpec struct {
	Version   string                     `json:"version"`
	Plugins   []CouchbaseClusterPlugin   `json:"plugins"`
	NodePools []CouchbaseClusterNodePool `json:"nodePools"`
	// TODO: Remove this field
	Image  CouchbaseImage `json:"image"`
	Sysctl []string       `json:"sysctl"`
}

// CouchbaseClusterPlugin describes a specification of an CouchbaseCluster plugin
// You must ensure the plugin is compatible with the version of Couchbase being deployed
// else the cluster will not deploy successfully
type CouchbaseClusterPlugin struct {
	Name string `json:"name"`
}

// CouchbaseClusterNodePool describes a node pool within an CouchbaseCluster.
// The nodes in this pool will be configured to be of the specified roles
type CouchbaseClusterNodePool struct {
	Name        string                             `json:"name"`
	Replicas    int32                              `json:"replicas"`
	Roles       []string                           `json:"roles"`
	Resources   *v1.ResourceRequirements           `json:"resources,omitempty"`
	Persistence *CouchbaseClusterPersistenceConfig `json:"persistence,omitempty"`
}

type CouchbaseClusterPersistenceConfig struct {
	Size         string `json:"size"`
	StorageClass string `json:"storageClass"`
}

// TODO: Remove this struct
type CouchbaseImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
	FsGroup    int64  `json:"fsGroup"`
}
