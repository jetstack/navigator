package cassandra_test

import (
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	navigatorclientset "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	cluster *v1alpha1.CassandraCluster
	client  navigatorclientset.Interface
	control cassandra.ControlInterface
}

func setup(t *testing.T) *fixture {
	namespace := "foo"
	name := "bar"
	cluster := &v1alpha1.CassandraCluster{}
	cluster.SetName(name)
	cluster.SetNamespace(namespace)
	client := fake.NewSimpleClientset(cluster)
	control := cassandra.NewController(client)
	return &fixture{cluster, client, control}
}

func TestControl(t *testing.T) {
	t.Run(
		"new cluster",
		func(t *testing.T) {
			fixture := setup(t)
			err := fixture.control.Sync(fixture.cluster)
			assert.NoError(t, err)
		},
	)
}
