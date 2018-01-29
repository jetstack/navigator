package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

type esClusterBuilder struct {
	esc *v1alpha1.ElasticsearchCluster
}

func NewESCluster(name, namespace string) *esClusterBuilder {
	return &esClusterBuilder{
		esc: &v1alpha1.ElasticsearchCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}

// AddNodePool adds the given node pools to this cluster
func (e *esClusterBuilder) AddNodePool(np ...v1alpha1.ElasticsearchClusterNodePool) *esClusterBuilder {
	e.esc.Spec.NodePools = append(e.esc.Spec.NodePools, np...)
	return e
}

// Pilots returns the fully specified Pilots for the ElasticsearchCluster being
// built. This is useful when building a unit test against a cluster that
// requires Pilots to be present for the cluster.
func (e *esClusterBuilder) Pilots() []*v1alpha1.Pilot {

}
