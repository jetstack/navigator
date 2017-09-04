package nodepool

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/util"
)

func selectorForNodePool(c v1alpha1.ElasticsearchCluster, np v1alpha1.ElasticsearchClusterNodePool) (labels.Selector, error) {
	nodePoolNameReq, err := labels.NewRequirement(util.NodePoolNameLabelKey, selection.Equals, []string{np.Name})
	if err != nil {
		return nil, err
	}
	clusterSelector, err := selectorForCluster(c)
	if err != nil {
		return nil, err
	}
	return clusterSelector.Add(*nodePoolNameReq), nil
}

func selectorForCluster(c v1alpha1.ElasticsearchCluster) (labels.Selector, error) {
	clusterNameReq, err := labels.NewRequirement(util.NodePoolClusterLabelKey, selection.Equals, []string{c.Name})
	if err != nil {
		return nil, err
	}
	return labels.NewSelector().Add(*clusterNameReq), nil
}
