package e2e

import (
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jetstack/navigator/test/e2e/framework"
	// test sources
	_ "github.com/jetstack/navigator/test/e2e/elasticsearch"
)

func init() {
	framework.ViperizeFlags()
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}
