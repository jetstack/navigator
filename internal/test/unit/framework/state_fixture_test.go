package framework_test

import (
	"testing"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func TestFixturePopulatesConfigMapLister(t *testing.T) {
	fixture := &framework.StateFixture{
		T: t,
		KubeObjects: []runtime.Object{
			&core.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
	}
	fixture.Start()

	state := fixture.State()
	cms, err := state.ConfigMapLister.List(labels.Everything())
	if err != nil {
		t.Errorf("Got an error when listing ConfigMaps: %s", err.Error())
	}

	if len(cms) != 1 {
		t.Errorf("Expected 1 ConfigMaps to be returned from lister but got %d", len(cms))
	}
}

func TestFixturePopulatesPilotLister(t *testing.T) {
	fixture := &framework.StateFixture{
		T: t,
		NavigatorObjects: []runtime.Object{
			&v1alpha1.Pilot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
	}
	fixture.Start()

	state := fixture.State()
	pilots, err := state.PilotLister.List(labels.Everything())
	if err != nil {
		t.Errorf("Got an error when listing Pilots: %s", err.Error())
	}

	if len(pilots) != 1 {
		t.Errorf("Expected 1 Pilots to be returned from lister but got %d", len(pilots))
	}
}

func TestDeleteResource(t *testing.T) {
	fixture := &framework.StateFixture{
		NavigatorObjects: []runtime.Object{
			&v1alpha1.Pilot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
	}
	fixture.Start()

	state := fixture.State()
	pilots, err := state.PilotLister.List(labels.Everything())
	if err != nil {
		t.Errorf("Got an error when listing Pilots: %s", err.Error())
	}

	if len(pilots) != 1 {
		t.Errorf("Expected 1 Pilots to be returned from lister but got %d", len(pilots))
	}

	pilot := pilots[0]
	err = state.NavigatorClientset.NavigatorV1alpha1().Pilots(pilot.Namespace).Delete(pilot.Name, nil)
	if err != nil {
		t.Errorf("Failed to delete test Pilot %q", pilot.Name)
	}

	// force pilot controller to resync
	fixture.WaitForResync()

	// TODO: This check will currently fail due to needing #57504 from k/k.
	// For the meantime, we need to be aware of this when writing our tests and also
	// only use the Clientset to check the contents of the API when running assertions.
	// pilots, err = state.PilotLister.List(labels.Everything())
	// if err != nil {
	// 	t.Errorf("Error listing pilots: %v", err)
	// }
	// if len(pilots) != 0 {
	// 	t.Errorf("Expected pilot to be deleted, but Pilot is still present in lister: %v", pilots)
	// }

	pilotList, err := state.NavigatorClientset.NavigatorV1alpha1().Pilots(pilot.Namespace).List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Error listing pilots: %v", err)
	}
	if len(pilotList.Items) != 0 {
		t.Errorf("Expected pilot to be deleted, but Pilot is still present in apiserver: %v", pilotList.Items)
	}
}
