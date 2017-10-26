package cassandra

import (
	"testing"
	"time"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigatorfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/watch"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type fixture struct {
	t          *testing.T
	nclient    *navigatorfake.Clientset
	nwatch     *watch.FakeWatcher
	nfactory   externalversions.SharedInformerFactory
	recorder   *record.FakeRecorder
	controller *CassandraController
	finished   chan struct{}
	stopCh     chan struct{}
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
	recorder := record.NewFakeRecorder(0)

	controller := NewCassandra(nclient, nil, cassClusters, recorder)

	return &fixture{
		t:          t,
		nclient:    nclient,
		nwatch:     nwatch,
		nfactory:   nfactory,
		recorder:   recorder,
		controller: controller,
	}
}

func (f *fixture) run() {
	f.stopCh = make(chan struct{})
	f.finished = make(chan struct{})
	f.nfactory.Start(f.stopCh)
	go func() {
		defer close(f.finished)
		err := f.controller.Run(1, f.stopCh)
		if err != nil {
			f.t.Error(err)
		}
	}()

	if !cache.WaitForCacheSync(
		f.stopCh,
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
	if f.stopCh != nil {
		f.nwatch.Add(cluster)
	}
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

func (f *fixture) expectEvent() {
	select {
	case e := <-f.recorder.Events:
		f.t.Log(e)
	case <-time.After(time.Second):
		f.t.Error("Timed out waiting for event")
	}
}

func (f *fixture) close() {
	close(f.stopCh)
	<-f.finished
}

func newCassandraCluster() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{}
	c.SetName("cassandra-1")
	c.SetNamespace("app-1")
	return c
}

func TestCassandraController(t *testing.T) {
	t.Run(
		"CassandraController syncs in response to informer events",
		func(t *testing.T) {
			f := newFixture(t)
			cluster := newCassandraCluster()
			f.run()
			defer f.close()
			f.add(cluster)
			f.expectEvent()
		},
	)
	t.Run(
		"CassandraController handles deleted CassandraCluster resources",
		func(t *testing.T) {
			f := newFixture(t)
			cluster := newCassandraCluster()
			f.run()
			defer f.close()
			f.add(cluster)
			f.expectEvent()
			f.delete(cluster)
			// XXX This stinks, but I haven't got another way to know when the
			// controller has responded to the delete event from the informer.
			// In followup branches I'll call CassandraCluster.sync directly, to
			// avoid this async test.
			<-time.After(time.Second)
		},
	)
}
