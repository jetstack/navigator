package elasticsearch

import "gitlab.jetstack.net/marshal/colonel/pkg/api/v1"

func (e *ElasticsearchController) clusterNeedsUpdate(es *v1.ElasticsearchCluster) (bool, error) {
	for _, np := range es.Spec.NodePools {
		if needsUpdate, err := e.nodePoolNeedsUpdate(es, np); err != nil {
			return false, err
		} else if needsUpdate {
			return true, nil
		}
	}
	return false, nil
}
