package rolebinding_test

import (
	"fmt"
	"testing"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/rolebinding"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestRoleBindingSync(t *testing.T) {
	t.Run(
		"role missing",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertRoleBindingsLength(1)
		},
	)
	t.Run(
		"role exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			existingRoleBinding := rolebinding.RoleBindingForCluster(f.Cluster)
			f.AddObjectK(existingRoleBinding)
			f.Run()
			f.AssertRoleBindingsLength(1)
		},
	)
	t.Run(
		"sync fails",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.RoleBindingControl = &casstesting.FakeControl{
				SyncError: fmt.Errorf("simulated sync error"),
			}
			f.RunExpectError()
		},
	)
}
