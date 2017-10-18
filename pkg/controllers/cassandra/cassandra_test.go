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
	cc := cassandra.NewCassandra(clientset, i)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		err := cc.Run(2, stopCh)
		if err != nil {
			t.Error(err)
		}
	}()

	defer func() {
		close(stopCh)
		<-finished
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
			_, err := clientset.NavigatorV1alpha1().CassandraClusters(c.Namespace).Create(c)
			if err != nil {
				t.Fatal(err)
			}
			fakeWatch.Add(c)
			<-time.After(time.Second)
		},
	)
	t.Run(
		"Delete a cluster",
		func(t *testing.T) {
			err := clientset.NavigatorV1alpha1().CassandraClusters(c.Namespace).Delete(c.Name, nil)
			if err != nil {
				t.Fatal(err)
			}
			fakeWatch.Delete(c)
			<-time.After(time.Second)
		},
	)
	t.Run(
		"Simulate a sync error",
		func(t *testing.T) {
			// Don't add the cluster to the clientset. The call to UpdateStatus will fail.
			fakeWatch.Add(c)
			<-time.After(time.Second)
		},
	)
}
