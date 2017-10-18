package cassandra_test

import (
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	navigatorclientset "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	cluster *v1alpha1.CassandraCluster
	nclient navigatorclientset.Interface
	kclient kubernetes.Interface
	control cassandra.ControlInterface
}

func setup(t *testing.T) *fixture {
	namespace := "foo"
	name := "bar"
	cluster := &v1alpha1.CassandraCluster{}
	cluster.SetName(name)
	cluster.SetNamespace(namespace)
	kclient := fake.NewSimpleClientset()
	nclient := navigatorfake.NewSimpleClientset(cluster)
	control := cassandra.NewController(nclient, kclient)
	return &fixture{cluster, nclient, kclient, control}
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
