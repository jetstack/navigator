package cassandra_test

import (
	"testing"
	"time"

	informerv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions/navigator/v1alpha1"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra"
	"k8s.io/apimachinery/pkg/watch"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

func TestCassandraController(t *testing.T) {
	namespace := "foo"
	name := "bar"
	c := &v1alpha1.CassandraCluster{}
	c.SetName(name)
	c.SetNamespace(namespace)

	stopCh := make(chan struct{})

	clientset := fake.NewSimpleClientset()
	fakeWatch := watch.NewFake()
	clientset.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(fakeWatch, nil),
	)

	i := informerv1alpha1.NewCassandraClusterInformer(
		clientset,
		namespace,
		0,
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		},
	)
	go i.Run(stopCh)
	cc := cassandra.NewCassandra(i)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		err := cc.Run(2, stopCh)
		if err != nil {
			t.Error(err)
		}
	}()
	if !cache.WaitForCacheSync(
		stopCh,
		i.HasSynced,
	) {
		t.Errorf("timed out waiting for caches to sync")
	}
	t.Run(
		"Create a cluster",
		func(t *testing.T) {
			fakeWatch.Add(c)
			<-time.After(time.Second)
		},
	)
	t.Run(
		"Delete a cluster",
		func(t *testing.T) {
			fakeWatch.Delete(c)
			<-time.After(time.Second)
		},
	)
	close(stopCh)
	<-finished
}
