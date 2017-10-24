package cassandra_test

import (
	"testing"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
)

type fixture struct {
	cluster *v1alpha1.CassandraCluster
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
	kubeFactory := informers.NewSharedInformerFactory(kclient, 0)
	serviceLister := kubeFactory.Core().V1().Services().Lister()
	control := cassandra.NewControl(
		service.NewControl(kclient, serviceLister),
	)
	return &fixture{cluster, kclient, control}
}

func TestControl(t *testing.T) {
	t.Run(
		"new cluster",
		func(t *testing.T) {
			fixture := setup(t)
			err := fixture.control.Sync(fixture.cluster)
			if err != nil {
				t.Error(err)
			}
		},
	)
}
