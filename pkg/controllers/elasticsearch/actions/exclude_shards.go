package actions

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
)

type ExcludeShards struct {
	// The Elasticsearch cluster resource being operated on
	Cluster *v1alpha1.ElasticsearchCluster
	// The node pool being scaled
	NodePool *v1alpha1.ElasticsearchClusterNodePool
	// Number of replicas to scale to
	Replicas int32
}

var _ controllers.Action = &ExcludeShards{}

func (c *ExcludeShards) Name() string {
	return "ExcludeShards"
}

func (c *ExcludeShards) Message() string {
	return ""
}

func (e *ExcludeShards) Execute(state *controllers.State) error {
	return nil
}
