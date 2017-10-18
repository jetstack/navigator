package service_test

import (
	"testing"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
	"github.com/stretchr/testify/assert"
)

func TestServiceControl(t *testing.T) {
	cluster := &v1alpha1.CassandraCluster{}
	cluster.SetNamespace("foo")
	cluster.SetName("bar")
	kclient := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(kclient, 0)
	serviceLister := factory.Core().V1().Services().Lister()
	c := service.NewControl(kclient, serviceLister)
	err := c.Sync(cluster)
	assert.NoError(t, err)
}
