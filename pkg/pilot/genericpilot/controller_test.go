package genericpilot

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
	informers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/process"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/scheduler"
)

type fixture struct {
	// the name of the test
	name string
	// the key to pass to sync
	key string
	// whether to expect an error to be returned
	err bool
	// the 'thisPilot' name & namespace options for the pilot
	thisPilotName, thisPilotNamespace string
	// items expected to be scheduled to be re-run
	expectedScheduledItems []interface{}
	// actions expected to happen on the client.
	navActions []core.Action
	// contents of the pilotLister
	pilotLister []*v1alpha1.Pilot
	// list of objects to load into the clientset
	kubeObjects []runtime.Object
	// optional hooks resource for testing
	hooks *hook.Hooks
	// cachedThisPilot field on the controller. useful for testing that
	// 'offline mode' works
	cachedThisPilot *v1alpha1.Pilot
	// a fake process for the controller
	process process.Interface
	// if set, this fixture will ensure the Pilot passed to the controllers
	// SyncFunc deep equals this pilot
	syncFuncExpectedPilot *v1alpha1.Pilot
}

func (f *fixture) run(t *testing.T) {
	// initialise hooks if nil
	if f.hooks == nil {
		f.hooks = &hook.Hooks{}
	}
	// set up testing scheduled workqueue
	expectedScheduledItems := make([]interface{}, len(f.expectedScheduledItems))
	copy(expectedScheduledItems, f.expectedScheduledItems)
	testScheduledWorkQueue := &scheduler.FakeScheduledWorkQueue{
		AddFunc: func(i interface{}, _ time.Duration) {
			for index, element := range expectedScheduledItems {
				if reflect.DeepEqual(i, element) {
					expectedScheduledItems = append(expectedScheduledItems[:index], expectedScheduledItems[index+1:]...)
					return
				}
			}
			t.Errorf("Unexpected item scheduled: %#v", i)
		},
		ForgetFunc: func(i interface{}) {
			// do nothing
		},
	}
	// set up clientset and informers
	navClient := navfake.NewSimpleClientset(f.kubeObjects...)
	navSIF := informers.NewSharedInformerFactory(navClient, 0)
	// fill pilotLister with items
	for _, pilot := range f.pilotLister {
		navSIF.Navigator().V1alpha1().Pilots().Informer().GetIndexer().Add(pilot)
	}
	// print event recorder events
	// TODO: test these too
	fakeRecorder := record.NewFakeRecorder(0)
	defer close(fakeRecorder.Events)
	go func() {
		for e := range fakeRecorder.Events {
			t.Logf("Event logged: %s", e)
		}
	}()
	// create testing sync func
	syncFunc := func(p *v1alpha1.Pilot) error {
		if f.syncFuncExpectedPilot == nil {
			return nil
		}
		if !reflect.DeepEqual(p, f.syncFuncExpectedPilot) {
			t.Errorf("Pilot passed to SyncFunc %#v does not match expected pilot %#v", p, f.syncFuncExpectedPilot)
		}
		return nil
	}

	// create genericpilot
	g := &GenericPilot{
		Options: Options{
			PilotNamespace: f.thisPilotNamespace,
			PilotName:      f.thisPilotName,
			Hooks:          f.hooks,
			SyncFunc:       syncFunc,
		},
		cachedThisPilot:    f.cachedThisPilot,
		scheduledWorkQueue: testScheduledWorkQueue,
		pilotLister:        navSIF.Navigator().V1alpha1().Pilots().Lister(),
		client:             navClient,
		process:            f.process,
		recorder:           fakeRecorder,
	}

	// run sync
	err := g.sync(f.key)
	if err != nil && !f.err {
		t.Errorf("Unexpected error from Sync for key %q: %s", f.key, err.Error())
		return
	}

	// the fake scheduler above remove elements from this slice, so if any are
	// left here then an item that should have been scheduled was not
	if len(expectedScheduledItems) > 0 {
		t.Errorf("Expected scheduled items %#v to be scheduled but not found in scheduled queue", expectedScheduledItems)
		return
	}

	// get a list of actions performed on the navigator client
	navActions := navClient.Actions()
	for i, action := range navActions {
		// check for unexpected actions
		if len(f.navActions) < i+1 {
			t.Errorf("%d unexpected actions: %+v", len(navActions)-len(f.navActions), navActions[i:])
			break
		}

		expectedAction := f.navActions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) && action.GetSubresource() == expectedAction.GetSubresource()) {
			t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.navActions) > len(navActions) {
		t.Errorf("%d additional expected actions:%+v", len(f.navActions)-len(navActions), f.navActions[len(navActions):])
	}
}

func newUpdatePilotStatusAction(p *v1alpha1.Pilot) core.Action {
	a := core.NewUpdateAction(schema.GroupVersionResource{
		Group:    navigator.GroupName,
		Resource: "pilots",
		Version:  "v1alpha1",
	}, p.Namespace, p)
	a.Subresource = "status"
	return a
}

// returns a 'default' process adapter that always returns true for Running()
func newDefaultProcessAdapter() process.Interface {
	return &process.FakeAdapter{
		RunningFunc: func() bool { return true },
		StringFunc:  func() string { return "ps-adapter" },
	}
}

func TestSync(t *testing.T) {
	tests := []*fixture{
		{
			name:                   "successful run of Sync on 'this' pilot should return no error and update status",
			key:                    "namespace/name",
			err:                    false,
			thisPilotName:          "name",
			thisPilotNamespace:     "namespace",
			expectedScheduledItems: []interface{}{"namespace/name"},
			pilotLister:            []*v1alpha1.Pilot{newDummyPilot("name", "namespace")},
			kubeObjects:            []runtime.Object{newDummyPilot("name", "namespace")},
			navActions:             []core.Action{newUpdatePilotStatusAction(newDummyPilot("name", "namespace"))},
			process:                newDefaultProcessAdapter(),
		},
		{
			name:                   "should not Start a subprocess when syncing a non-this pilot",
			key:                    "namespace/not-this-pilot",
			err:                    false,
			thisPilotName:          "name",
			thisPilotNamespace:     "namespace",
			expectedScheduledItems: []interface{}{"namespace/not-this-pilot"},
			cachedThisPilot:        newDummyPilot("name", "namespace"),
			pilotLister:            []*v1alpha1.Pilot{newDummyPilot("not-this-pilot", "namespace")},
			kubeObjects:            []runtime.Object{newDummyPilot("not-this-pilot", "namespace")},
		},
		{
			name:                   "failed genericpilot consumer sync function should fail sync",
			key:                    "namespace/name",
			err:                    true,
			thisPilotName:          "name",
			thisPilotNamespace:     "namespace",
			expectedScheduledItems: []interface{}{"namespace/name"},
			pilotLister:            []*v1alpha1.Pilot{newDummyPilot("name", "namespace")},
			kubeObjects:            []runtime.Object{newDummyPilot("name", "namespace")},
			navActions:             []core.Action{newUpdatePilotStatusAction(newDummyPilot("name", "namespace"))},
			process:                newDefaultProcessAdapter(),
		},
		{
			name:               "missing cachedThisPilot and empty clientset should fail Sync()",
			key:                "namespace/name",
			err:                true,
			thisPilotName:      "name",
			thisPilotNamespace: "namespace",
		},
		{
			name: "should reuse cached 'this' pilot resource, and should fail to update status",
			key:  "namespace/name",
			err:  true,
			syncFuncExpectedPilot:  newDummyPilot("name", "namespace"),
			thisPilotName:          "name",
			thisPilotNamespace:     "namespace",
			expectedScheduledItems: []interface{}{"namespace/name"},
			cachedThisPilot:        newDummyPilot("name", "namespace"),
			navActions:             []core.Action{newUpdatePilotStatusAction(newDummyPilot("name", "namespace"))},
			process:                newDefaultProcessAdapter(),
		},
		{
			name: "should sync not-this-pilot and should not update the pilots status",
			key:  "namespace/not-this-pilot",
			err:  false,
			syncFuncExpectedPilot:  newDummyPilot("not-this-pilot", "namespace"),
			thisPilotName:          "name",
			thisPilotNamespace:     "namespace",
			expectedScheduledItems: []interface{}{"namespace/not-this-pilot"},
			cachedThisPilot:        newDummyPilot("name", "namespace"),
			pilotLister:            []*v1alpha1.Pilot{newDummyPilot("not-this-pilot", "namespace")},
			kubeObjects:            []runtime.Object{newDummyPilot("not-this-pilot", "namespace")},
			process:                newDefaultProcessAdapter(),
		},
	}
	// we can't use subtests yet as functions passed to our fixtures may need
	// to reference 't' themselves
	for _, test := range tests {
		t.Run(test.name, func(test *fixture) func(*testing.T) {
			return func(t *testing.T) {
				test.run(t)
			}
		}(test))
	}
}

func newDummyPilot(name, namespace string) *v1alpha1.Pilot {
	return &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apiversion",
					Kind:       "generickind",
					Name:       "name",
					UID:        "",
					Controller: util.BoolPtr(true),
				},
			},
		},
	}
}
