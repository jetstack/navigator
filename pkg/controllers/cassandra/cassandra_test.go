package cassandra_test

import (
	"strings"
	"testing"
	"time"

	navigatorfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
	"github.com/jetstack/navigator/pkg/kube"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
)

type fixture struct {
	t           *testing.T
	kclient     *fake.Clientset
	nclient     *navigatorfake.Clientset
	nwatch      *watch.FakeWatcher
	recorder    *record.FakeRecorder
	syncSuccess chan struct{}
	finished    chan struct{}
}

func NewFixture(t *testing.T) *fixture {
	nclient := navigatorfake.NewSimpleClientset()
	nwatch := watch.NewFake()
	nclient.PrependWatchReactor(
		"cassandraclusters",
		clienttesting.DefaultWatchReactor(nwatch, nil),
	)

	return &fixture{
		t:           t,
		kclient:     fake.NewSimpleClientset(),
		nclient:     nclient,
		nwatch:      nwatch,
		recorder:    record.NewFakeRecorder(0),
		syncSuccess: make(chan struct{}),
		finished:    make(chan struct{}),
	}
}

func (f *fixture) run() {
	go func() {
		for e := range f.recorder.Events {
			f.t.Logf("EVENT: %q", e)
			if strings.Contains(e, cassandra.SuccessSync) {
				close(f.syncSuccess)
			}
		}
		close(f.finished)
	}()
}

// NewCassandra sets up event handlers for the supplied informers.
func TestCassandraControllerIntegration(t *testing.T) {
	f := NewFixture(t)

	ctx := &controllers.Context{
		Client:                f.kclient,
		NavigatorClient:       f.nclient,
		Recorder:              f.recorder,
		SharedInformerFactory: kube.NewSharedInformerFactory(),
		Namespace:             "",
	}
	controller := cassandra.CassandraControllerFromContext(ctx)

	stopCh := make(chan struct{})
	ctx.SharedInformerFactory.Start(stopCh)
	controllerFinished := make(chan struct{})
	go func() {
		defer func() {
			close(f.recorder.Events)
			close(controllerFinished)
		}()
		err := controller.Run(1, stopCh)
		if err != nil {
			t.Error(err)
		}
	}()
	f.run()
	defer func() {
		close(stopCh)
		<-controllerFinished
		<-f.finished
	}()

	cluster := casstesting.ClusterForTest()
	f.nwatch.Add(cluster)
	select {
	case <-f.syncSuccess:
	case <-time.After(time.Second):
		t.Error("Timeout waiting for controller sync")
	}
}
