package cassandra_test

import (
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra"
)

func cluster(t *testing.T) *v1alpha1.CassandraCluster {
	namespace := "foo"
	name := "bar"
	cluster := &v1alpha1.CassandraCluster{}
	cluster.SetName(name)
	cluster.SetNamespace(namespace)
	return cluster
}

func TestControl(t *testing.T) {
	client := fake.NewSimpleClientset()
	c := cassandra.NewController(client)

	t.Run(
		"new cluster",
		func(t *testing.T) {
			cluster := cluster(t)
			err := c.Sync(cluster)
			if err != nil {
				t.Fatal(err)
			}
		},
	)
}
