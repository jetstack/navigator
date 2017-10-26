package service_test

import (
	"fmt"
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
	casstesting "github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/testing"
	"k8s.io/api/core/v1"
)

func TestServiceSync(t *testing.T) {
	t.Run(
		"sync error",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.ServiceControl = &casstesting.FakeControl{
				SyncError: fmt.Errorf("simulated sync error"),
			}
			f.RunExpectError()
		},
	)
	t.Run(
		"service created",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertServicesLength(1)
		},
	)
	t.Run(
		"service exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.AddObjectK(service.ServiceForCluster(f.Cluster))
			f.Run()
			f.AssertServicesLength(1)
		},
	)
	t.Run(
		"service needs sync",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			// Remove the ports from the default cluster and expect them to be
			// re-created.
			unsyncedService := service.ServiceForCluster(f.Cluster)
			unsyncedService.Spec.Ports = []v1.ServicePort{}
			f.AddObjectK(unsyncedService)
			f.Run()
			services := f.Services()
			service := services.Items[0]
			if len(service.Spec.Ports) == 0 {
				t.Log(service)
				t.Error("Service was not updated")
			}
		},
	)
}
