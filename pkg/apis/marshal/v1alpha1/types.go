package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// Required to satisfy Object interface
func (e *ElasticsearchCluster) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *ElasticsearchCluster) GetObjectMeta() metav1.Object {
	return &e.ObjectMeta
}

// Required to satisfy Object interface
func (el *ElasticsearchClusterList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *ElasticsearchClusterList) GetListMeta() metav1.List {
	return &el.ListMeta
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
	Name            string
	Replicas        int32
	Roles           []string
	Resources       *v1.ResourceRequirements
	State           *ElasticsearchClusterStateConfig
	OwnerReferences []*ElasticsearchOwnerReference
}

type ElasticsearchClusterStateConfig struct {
	Stateful    bool
	Persistence *ElasticsearchClusterPersistenceConfig
}

type ElasticsearchClusterPersistenceConfig struct {
	Enabled      bool
	Size         string
	StorageClass string
}

type ElasticsearchImage struct {
	Repository string
	Tag        string
	PullPolicy string
	FsGroup    int64
}

type ElasticsearchOwnerReference struct {
	ApiVersion string
	Controller string
	Kind       string
	Name       string
	Uid        string
}
