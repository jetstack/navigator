package cassandra

import (
	"strings"
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack-experimental/navigator/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
)

func newCassandraCluster() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{}
	c.SetName("cassandra-1")
	c.SetNamespace("app-1")
	return c
}

// A NewCassandra.sync is called in response to events from the supplied informers.
func TestCassandraControllerIntegration(t *testing.T) {
	nclient := navigatorfake.NewSimpleClientset()
	nwatch := watch.NewFake()
	nclient.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(nwatch, nil),
	)
	nfactory := externalversions.NewSharedInformerFactory(nclient, 0)
	cassClusters := nfactory.Navigator().V1alpha1().CassandraClusters().Informer()

	kclient := fake.NewSimpleClientset()
	kfactory := informers.NewSharedInformerFactory(kclient, 0)
	services := kfactory.Core().V1().Services().Informer()

	recorder := record.NewFakeRecorder(0)

	controller := NewCassandra(nclient, kclient, cassClusters, services, recorder)

	stopCh := make(chan struct{})
	nfactory.Start(stopCh)
	kfactory.Start(stopCh)
	controllerFinished := make(chan struct{})
	syncFinished := make(chan struct{})
	go func() {
		for e := range recorder.Events {
			t.Logf("EVENT: %q", e)
			if strings.Contains(e, messageSuccessSync) {
				close(syncFinished)
			}
		}
	}()
	go func() {
		defer close(controllerFinished)
		err := controller.Run(1, stopCh)
		if err != nil {
			t.Error(err)
		}
	}()
	defer func() {
		<-syncFinished
		close(stopCh)
		<-controllerFinished
	}()

	cluster := newCassandraCluster()
	nwatch.Add(cluster)
}
