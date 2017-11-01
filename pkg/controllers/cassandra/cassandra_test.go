package cassandra_test

import (
	"strings"
	"testing"

	navigatorfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
)

// NewCassandra sets up event handlers for the supplied informers.
func TestCassandraControllerIntegration(t *testing.T) {
	nclient := navigatorfake.NewSimpleClientset()
	nwatch := watch.NewFake()
	nclient.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(nwatch, nil),
	)
	nfactory := externalversions.NewSharedInformerFactory(nclient, 0)
	cassClusters := nfactory.Navigator().V1alpha1().CassandraClusters()

	kclient := fake.NewSimpleClientset()
	kfactory := informers.NewSharedInformerFactory(kclient, 0)
	services := kfactory.Core().V1().Services()

	recorder := record.NewFakeRecorder(0)

	controller := cassandra.NewCassandra(nclient, kclient, cassClusters, services, recorder)

	stopCh := make(chan struct{})
	nfactory.Start(stopCh)
	kfactory.Start(stopCh)
	controllerFinished := make(chan struct{})
	syncFinished := make(chan struct{})
	go func() {
		for e := range recorder.Events {
			t.Logf("EVENT: %q", e)
			if strings.Contains(e, cassandra.MessageSuccessSync) {
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

	cluster := casstesting.ClusterForTest()
	nwatch.Add(cluster)
}
