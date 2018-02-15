package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func ClusterLabels(c *v1alpha1.ElasticsearchCluster) map[string]string {
	return map[string]string{
		v1alpha1.ElasticsearchClusterNameLabel: c.Name,
	}
}

func NodePoolLabels(c *v1alpha1.ElasticsearchCluster, poolName string, roles ...v1alpha1.ElasticsearchClusterRole) map[string]string {
	labels := ClusterLabels(c)
	if poolName != "" {
		labels[v1alpha1.ElasticsearchNodePoolNameLabel] = poolName
	}
	for _, role := range roles {
		labels[RoleLabel(role)] = "true"
	}
	return labels
}

func RoleLabel(role v1alpha1.ElasticsearchClusterRole) string {
	return v1alpha1.ElasticsearchRoleLabelPrefix + string(role)
}

func NodePoolResourceName(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) string {
	return fmt.Sprintf("%s-%s", ResourceBaseName(c), np.Name)
}

func SelectorForNodePool(clusterName, poolName string) (labels.Selector, error) {
	nodePoolNameReq, err := labels.NewRequirement(v1alpha1.ElasticsearchNodePoolNameLabel, selection.Equals, []string{poolName})
	if err != nil {
		return nil, err
	}
	clusterSelector, err := SelectorForCluster(clusterName)
	if err != nil {
		return nil, err
	}
	return clusterSelector.Add(*nodePoolNameReq), nil
}
