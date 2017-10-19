package cassandra

import (
	"testing"
	"time"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	"github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
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

	nclient := navigatorfake.NewSimpleClientset()
	fakeWatch := watch.NewFake()
	nclient.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(fakeWatch, nil),
	)
	naviFactory := externalversions.NewSharedInformerFactory(nclient, 0)
	go naviFactory.Start(stopCh)

	kclient := fake.NewSimpleClientset()
	kubeFactory := informers.NewSharedInformerFactory(kclient, 0)
	go kubeFactory.Start(stopCh)

	cc := NewCassandra(
		nclient,
		kclient,
		naviFactory.Navigator().V1alpha1().CassandraClusters().Informer(),
		kubeFactory.Core().V1().Services().Informer(),
	)
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
		cc.cassListerSynced, cc.servicesListerSynced,
	) {
		t.Errorf("timed out waiting for caches to sync")
	}

	defer func() {
		close(stopCh)
		<-finished
	}()

	t.Run(
		"Create a cluster",
		func(t *testing.T) {
			_, err := nclient.NavigatorV1alpha1().CassandraClusters(c.Namespace).Create(c)
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
			err := nclient.NavigatorV1alpha1().CassandraClusters(c.Namespace).Delete(c.Name, nil)
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
