package role_test

import (
	"fmt"
	"testing"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/role"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestRoleSync(t *testing.T) {
	t.Run(
		"role missing",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertRolesLength(1)
		},
	)
	t.Run(
		"role exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			existingRole := role.RoleForCluster(f.Cluster)
			f.AddObjectK(existingRole)
			f.Run()
			f.AssertRolesLength(1)
		},
	)
	t.Run(
		"sync fails",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.RoleControl = &casstesting.FakeControl{
				SyncError: fmt.Errorf("simulated sync error"),
			}
			f.RunExpectError()
		},
	)
}
