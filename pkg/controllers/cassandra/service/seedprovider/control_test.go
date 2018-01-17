package seedprovider_test

import (
	"fmt"
	"testing"

	apiv1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/service/seedprovider"
	servicetesting "github.com/jetstack/navigator/pkg/controllers/cassandra/service/testing"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func newFixture(t *testing.T) *casstesting.Fixture {
	f := casstesting.NewFixture(t)
	f.CqlServiceControl = &casstesting.FakeControl{}
	return f
}

func newService(f *casstesting.Fixture) *apiv1.Service {
	return seedprovider.ServiceForCluster(f.Cluster)
}

func TestSeedProviderServiceSync(t *testing.T) {
	servicetesting.RunStandardServiceTests(t, newFixture, newService)
	t.Run(
		"sync error",
		func(t *testing.T) {
			f := newFixture(t)
			f.SeedProviderServiceControl = &casstesting.FakeControl{
				SyncError: fmt.Errorf("simulated sync error"),
			}
			f.RunExpectError()
		},
	)
}
