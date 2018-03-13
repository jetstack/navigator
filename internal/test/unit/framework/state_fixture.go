package framework

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"

	navfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	informers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack/navigator/pkg/controllers"
)

const informerResyncPeriod = time.Millisecond * 500

type StateFixture struct {
	T                *testing.T
	KubeObjects      []runtime.Object
	NavigatorObjects []runtime.Object

	kubeClient      *kubefake.Clientset
	navigatorClient *navfake.Clientset

	stopCh                         chan struct{}
	kubeSharedInformerFactory      kubeinformers.SharedInformerFactory
	navigatorSharedInformerFactory informers.SharedInformerFactory

	state *controllers.State
}

func (s *StateFixture) State() *controllers.State {
	return s.state
}

func (s *StateFixture) KubeClient() *kubefake.Clientset {
	return s.kubeClient
}

func (s *StateFixture) NavigatorClient() *navfake.Clientset {
	return s.navigatorClient
}

func (s *StateFixture) Start() {
	s.kubeClient = kubefake.NewSimpleClientset(s.KubeObjects...)
	s.navigatorClient = navfake.NewSimpleClientset(s.NavigatorObjects...)

	s.kubeSharedInformerFactory = kubeinformers.NewSharedInformerFactory(s.kubeClient, informerResyncPeriod)
	s.navigatorSharedInformerFactory = informers.NewSharedInformerFactory(s.navigatorClient, informerResyncPeriod)
	s.state = &controllers.State{
		Clientset:          s.kubeClient,
		NavigatorClientset: s.navigatorClient,
		Recorder:           record.NewFakeRecorder(5),
		StatefulSetLister:  s.kubeSharedInformerFactory.Apps().V1beta1().StatefulSets().Lister(),
		ConfigMapLister:    s.kubeSharedInformerFactory.Core().V1().ConfigMaps().Lister(),
		PilotLister:        s.navigatorSharedInformerFactory.Navigator().V1alpha1().Pilots().Lister(),
		PodLister:          s.kubeSharedInformerFactory.Core().V1().Pods().Lister(),
		ServiceLister:      s.kubeSharedInformerFactory.Core().V1().Services().Lister(),
	}
	s.stopCh = make(chan struct{})
	s.kubeSharedInformerFactory.Start(s.stopCh)
	s.navigatorSharedInformerFactory.Start(s.stopCh)
	if err := mustAllSync(s.kubeSharedInformerFactory.WaitForCacheSync(s.stopCh)); err != nil {
		s.T.Fatalf("Error waiting for kubeSharedInformerFactory to sync: %v", err)
	}
	if err := mustAllSync(s.navigatorSharedInformerFactory.WaitForCacheSync(s.stopCh)); err != nil {
		s.T.Fatalf("Error waiting for navigatorSharedInformerFactory to sync: %v", err)
	}
}

// Stop will signal the informers to stop watching changes
func (s *StateFixture) Stop() {
	close(s.stopCh)
}

// WaitForResync will wait for the informer factory informer duration by
// calling time.Sleep. This will ensure that all informer Stores are up to date
// with current information from the fake clients.
func (s *StateFixture) WaitForResync() {
	// add 100ms here to try and cut down on flakes
	time.Sleep(informerResyncPeriod + time.Millisecond*100)
}

func mustAllSync(in map[reflect.Type]bool) error {
	var errs []error
	for t, started := range in {
		if !started {
			errs = append(errs, fmt.Errorf("informer for %v not synced", t))
		}
	}
	return utilerrors.NewAggregate(errs)
}
