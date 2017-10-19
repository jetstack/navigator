package service_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
)

type fixture struct {
	t       *testing.T
	control service.Interface
	kclient *fake.Clientset
}

func newFixture(t *testing.T) *fixture {
	kclient := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(kclient, 0)
	serviceLister := factory.Core().V1().Services().Lister()
	control := service.NewControl(kclient, serviceLister)

	return &fixture{
		t:       t,
		control: control,
		kclient: kclient,
	}
}

func (f *fixture) run(cluster *v1alpha1.CassandraCluster) {
	err := f.control.Sync(cluster)
	if err != nil {
		f.t.Error(err)
	}
}

func (f *fixture) expectService(namespace, name string) {
	_, err := f.kclient.CoreV1().Services(namespace).Get(
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		f.t.Log("Actions:")
		f.t.Log(f.kclient.Actions())
		f.t.Error(err)
	}
}

func newCassandraCluster() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{}
	c.SetNamespace("foo")
	c.SetName("bar")
	return c
}

func TestServiceControl(t *testing.T) {
	t.Run(
		"service created",
		func(t *testing.T) {
			cluster := newCassandraCluster()
			f := newFixture(t)
			f.run(cluster)
			f.expectService(cluster.Namespace, cluster.Name+"-service")
		},
	)
	t.Run(
		"resync",
		func(t *testing.T) {
			cluster := newCassandraCluster()
			f := newFixture(t)
			f.run(cluster)
			f.expectService(cluster.Namespace, cluster.Name+"-service")
			f.run(cluster)
			f.expectService(cluster.Namespace, cluster.Name+"-service")
		},
	)
}
