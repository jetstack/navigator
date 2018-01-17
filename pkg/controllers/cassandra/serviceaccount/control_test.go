package serviceaccount_test

import (
	"fmt"
	"testing"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/serviceaccount"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestServiceAccountSync(t *testing.T) {
	t.Run(
		"service account missing",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertServiceAccountsLength(1)
		},
	)
	t.Run(
		"service account exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			existingServiceaccount := serviceaccount.ServiceAccountForCluster(f.Cluster)
			f.AddObjectK(existingServiceaccount)
			f.Run()
			f.AssertServiceAccountsLength(1)
		},
	)
	t.Run(
		"sync fails",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.ServiceAccountControl = &casstesting.FakeControl{
				SyncError: fmt.Errorf("simulated sync error"),
			}
			f.RunExpectError()
		},
	)
}
