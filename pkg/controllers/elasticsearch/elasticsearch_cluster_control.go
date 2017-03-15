package elasticsearch

import "gitlab.jetstack.net/marshal/colonel/pkg/api/v1"

type ElasticsearchClusterControl interface {
	SyncElasticsearchCluster(*v1.ElasticsearchCluster) error
}

type defaultElasticsearchClusterControl struct {
	nodePoolControl ElasticsearchClusterNodePoolControl
}

var _ ElasticsearchClusterControl = &defaultElasticsearchClusterControl{}

func NewElasticsearchClusterControl(
	nodePoolControl ElasticsearchClusterNodePoolControl,
) ElasticsearchClusterControl {
	return &defaultElasticsearchClusterControl{
		nodePoolControl: nodePoolControl,
	}
}

func (e *defaultElasticsearchClusterControl) SyncElasticsearchCluster(c *v1.ElasticsearchCluster) error {
	for _, np := range c.Spec.NodePools {
		// TODO: parallelise this?
		if err := e.nodePoolControl.SyncElasticsearchClusterNodePool(c, np); err != nil {
			return err
		}
	}
	return nil
}
