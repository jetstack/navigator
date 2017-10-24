package cassandra

import (
	"testing"
	"time"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack-experimental/navigator/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/watch"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

type fixture struct {
	t          *testing.T
	nclient    *navigatorfake.Clientset
	nwatch     *watch.FakeWatcher
	nfactory   externalversions.SharedInformerFactory
	controller *CassandraController
	finished   chan struct{}
}

func newFixture(t *testing.T) *fixture {
	nclient := navigatorfake.NewSimpleClientset()
	nwatch := watch.NewFake()
	nclient.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(nwatch, nil),
	)
	nfactory := externalversions.NewSharedInformerFactory(nclient, 0)
	cassClusters := nfactory.Navigator().V1alpha1().CassandraClusters().Informer()
	return &fixture{
		t:          t,
		nclient:    nclient,
		nwatch:     nwatch,
		nfactory:   nfactory,
		controller: NewCassandra(nclient, nil, cassClusters),
		finished:   make(chan struct{}),
	}
}

func (f *fixture) run(stopCh chan struct{}) {
	f.nfactory.Start(stopCh)
	go func() {
		defer close(f.finished)
		err := f.controller.Run(1, stopCh)
		if err != nil {
			f.t.Error(err)
		}
	}()

	if !cache.WaitForCacheSync(
		stopCh,
		f.controller.cassListerSynced,
	) {
		f.t.Errorf("timed out waiting for caches to sync")
	}
}

func (f *fixture) add(cluster *v1alpha1.CassandraCluster) {
	_, err := f.nclient.
		NavigatorV1alpha1().
		CassandraClusters(cluster.Namespace).
		Create(cluster)
	if err != nil {
		f.t.Fatal(err)
	}
	f.nwatch.Add(cluster)
}

func (f *fixture) delete(cluster *v1alpha1.CassandraCluster) {
	err := f.nclient.
		NavigatorV1alpha1().
		CassandraClusters(cluster.Namespace).
		Delete(cluster.Name, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	f.nwatch.Delete(cluster)
}

func newCassandraCluster() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{}
	c.SetName("cassandra-1")
	c.SetNamespace("app-1")
	return c
}

func TestCassandraController(t *testing.T) {
	f := newFixture(t)
	cluster := newCassandraCluster()
	stopCh := make(chan struct{})
	f.run(stopCh)
	defer func() {
		close(stopCh)
		<-f.finished
	}()

	t.Run(
		"Create a cluster",
		func(t *testing.T) {
			f.add(cluster)
			<-time.After(time.Second)
		},
	)
	t.Run(
		"Delete a cluster",
		func(t *testing.T) {
			f.delete(cluster)
			<-time.After(time.Second)
		},
	)
}
