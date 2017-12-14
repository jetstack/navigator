package pilot_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/jetstack/navigator/test/integration/framework"
)

func skipUnlessKubectlAvailable(t *testing.T) {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl not found")
	}
}

func TestESPilot(t *testing.T) {
	skipUnlessKubectlAvailable(t)
	t.Run(
		"pilot can be started and stopped, even in the absence of a corresponding Pilot resource.",
		func(t *testing.T) {
			f := framework.New(t)
			defer f.Wait()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			masterUrl := f.RunAMaster(ctx)
			esCtx, esCancel := context.WithCancel(ctx)
			finished := f.RunAnElasticSearchPilot(esCtx, masterUrl)
			esCancel()
			select {
			case <-finished:
			case <-time.After(time.Second * 10):
				close(finished)
				t.Fatal("ES pilot took longer than 10 seconds to stop")
			}
		},
	)
}
