package navigator

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// In this file we define the outer containing types for the CouchbaseCluster
// type. We could import these directly into message types defined in the types.proto
// file, but this is still TODO

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CouchbaseCluster describes a specification for an Couchbase cluster
type CouchbaseCluster struct {
	// we embed these types so the CouchbaseCluster implements runtime.Object
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   CouchbaseClusterSpec
	Status CouchbaseClusterStatus
}

type CouchbaseClusterStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CouchbaseClusterList defines a List type for our custom CouchbaseCluster type.
// This is needed in order to make List operations work.
type CouchbaseClusterList struct {
	// we embed these types so that CouchbaseClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta
	metav1.ListMeta

	Items []CouchbaseCluster
}

type CouchbaseClusterSpec struct {
	Version   string
	Plugins   []CouchbaseClusterPlugin
	NodePools []CouchbaseClusterNodePool
	Image     CouchbaseImage
	Sysctl    []string
}

type CouchbaseClusterPlugin struct {
	Name string
}

type CouchbaseClusterNodePool struct {
	Name        string
	Replicas    int32
	Roles       []string
	Resources   *v1.ResourceRequirements
	Persistence *CouchbaseClusterPersistenceConfig
}

type CouchbaseClusterPersistenceConfig struct {
	Size         string
	StorageClass string
}

type CouchbaseImage struct {
	Repository string
	Tag        string
	PullPolicy string
	FsGroup    int64
}
