package pilotcontroller_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kr/pretty"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigatorfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	informers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/pilotcontroller"
)

type fixture struct {
	t           *testing.T
	kclient     *fake.Clientset
	podWatch    *watch.FakeWatcher
	nclient     *navigatorfake.Clientset
	pilotWatch  *watch.FakeWatcher
	recorder    *record.FakeRecorder
	syncSuccess chan struct{}
	objectK     []runtime.Object
	objectN     []runtime.Object
}

func NewFixture(t *testing.T) *fixture {
	return &fixture{
		t:           t,
		recorder:    record.NewFakeRecorder(0),
		syncSuccess: make(chan struct{}),
	}
}

func (f *fixture) addObjectK(o ...runtime.Object) {
	f.objectK = append(f.objectK, o...)
}

func (f *fixture) addObjectN(o ...runtime.Object) {
	f.objectN = append(f.objectN, o...)
}

func (f *fixture) run() func() {
	f.kclient = fake.NewSimpleClientset(f.objectK...)
	f.podWatch = watch.NewFake()
	f.kclient.PrependWatchReactor(
		"pods",
		clienttesting.DefaultWatchReactor(f.podWatch, nil),
	)
	f.nclient = navigatorfake.NewSimpleClientset(f.objectN...)
	f.pilotWatch = watch.NewFake()
	f.nclient.PrependWatchReactor(
		"pilots",
		clienttesting.DefaultWatchReactor(f.pilotWatch, nil),
	)
	ctx := &controllers.Context{
		Client:                    f.kclient,
		NavigatorClient:           f.nclient,
		Recorder:                  f.recorder,
		KubeSharedInformerFactory: kubeinformers.NewSharedInformerFactory(f.kclient, 0),
		SharedInformerFactory:     informers.NewSharedInformerFactory(f.nclient, 0),
		Namespace:                 "namespace-not-used-in-this-test",
	}
	controller := pilotcontroller.NewFromContext(ctx)

	stopCh := make(chan struct{})
	ctx.SharedInformerFactory.Start(stopCh)
	ctx.KubeSharedInformerFactory.Start(stopCh)

	controllerFinished := make(chan struct{})
	go func() {
		defer func() {
			close(controllerFinished)
		}()
		err := controller.Run(1, stopCh)
		if err != nil {
			f.t.Error(err)
		}
	}()

	eventCheckerFinished := make(chan struct{})
	go func() {
		defer func() {
			close(eventCheckerFinished)
		}()
		for {
			select {
			case e := <-f.recorder.Events:
				f.t.Logf("EVENT: %q", e)
				if strings.Contains(e, pilotcontroller.SuccessSync) {
					f.syncSuccess <- struct{}{}
				}
			case <-stopCh:
				return
			}
		}
	}()
	return func() {
		close(stopCh)
		<-eventCheckerFinished
		<-controllerFinished
	}
}

func TestPilotControllerIntegration(t *testing.T) {
	cc1 := &v1alpha1.CassandraCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.Version,
			Kind:       "CassandraCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cc1",
			Namespace: "ns1",
		},
	}

	ss1 := &v1beta1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind: "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ss1",
			Namespace: "ns1",
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cc1, cc1.GroupVersionKind()),
			},
		},
	}

	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "p1",
			Namespace: "ns1",
			Labels: map[string]string{
				v1alpha1.PilotLabel: "",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ss1, ss1.GroupVersionKind()),
			},
		},
	}

	t.Run(
		"create pilot",
		func(t *testing.T) {
			f := NewFixture(t)
			// A pod exists without a corresponding pilot
			f.addObjectK(pod1)
			close := f.run()
			defer close()
			select {
			case <-f.syncSuccess:
			case <-time.After(time.Second):
				t.Fatal("Timeout waiting for controller sync")
			}
			pilots, err := f.nclient.NavigatorV1alpha1().Pilots("ns1").List(metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(pilots.Items) != 1 {
				t.Log(pretty.Sprint(pilots))
				t.Error("unexpected pilot count")
			}
		},
	)
}
