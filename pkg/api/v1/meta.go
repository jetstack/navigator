package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ElasticsearchCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ElasticsearchClusterSpec `json:"spec"`
}

type ElasticsearchClusterList struct {
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
