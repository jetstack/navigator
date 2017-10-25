package nodepool_test

import (
	"testing"

	casstesting "github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/testing"
)

func TestNodePoolControlSync(t *testing.T) {
	t.Run(
		"create a statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertStatefulSetsLength(1)
		},
	)
}
