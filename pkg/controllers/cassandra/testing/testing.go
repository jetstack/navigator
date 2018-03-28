package testing

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func ClusterForTest() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{
		Spec: v1alpha1.CassandraClusterSpec{
			NodePools: []v1alpha1.CassandraClusterNodePool{
				v1alpha1.CassandraClusterNodePool{
					Name:     "RingNodes",
					Replicas: 3,
				},
			},
		},
	}
	c.SetName("cassandra-1")
	c.SetNamespace("app-1")
	return c
}
