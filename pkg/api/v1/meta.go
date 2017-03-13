package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// In this file we define the outer containing types for the ElasticsearchCluster
// type. We could import these directly into message types defined in the types.proto
// file, but this is still TODO

// ElasticsearchCluster describes a specification for an Elasticsearch cluster
type ElasticsearchCluster struct {
	// we embed these types so the ElasticsearchCluster implements runtime.Object
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ElasticsearchClusterSpec `json:"spec"`
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
