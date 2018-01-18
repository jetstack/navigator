package testing

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"

	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

type FixtureFactory func(t *testing.T) *casstesting.Fixture
type ServiceFactory func(f *casstesting.Fixture) *apiv1.Service

func RunStandardServiceTests(t *testing.T, newFixture FixtureFactory, newService ServiceFactory) {
	t.Run(
		"service created",
		func(t *testing.T) {
			f := newFixture(t)
			f.Run()
			f.AssertServicesLength(1)
		},
	)
	t.Run(
		"service exists",
		func(t *testing.T) {
			f := newFixture(t)
			f.AddObjectK(newService(f))
			f.Run()
			f.AssertServicesLength(1)
		},
	)
	t.Run(
		"service needs sync",
		func(t *testing.T) {
			f := newFixture(t)
			// Remove the ports from the default cluster and expect them to be
			// re-created.
			unsyncedService := newService(f)
			unsyncedService.Spec.Selector = map[string]string{}
			f.AddObjectK(unsyncedService)
			f.Run()
			services := f.Services()
			service := services.Items[0]
			if len(service.Spec.Selector) == 0 {
				t.Log(service)
				t.Error("Service was not updated")
			}
		},
	)
	t.Run(
		"service with outside owner",
		func(t *testing.T) {
			f := newFixture(t)
			unownedService := newService(f)
			unownedService.OwnerReferences = nil
			f.AddObjectK(unownedService)
			f.RunExpectError()
		},
	)
}
