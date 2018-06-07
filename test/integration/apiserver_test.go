package integration_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jetstack/navigator/internal/test/integration/framework"
)

func TestAPIServerStandalone(t *testing.T) {
	cp := &framework.NavigatorControlPlane{
		NavigatorAPIServer: &framework.NavigatorAPIServer{
			Args: []string{
				"--standalone-mode",
			},
			Out: os.Stdout,
			Err: os.Stderr,
		},
	}
	err := cp.Start()
	require.NoError(t, err)
	defer func() {
		err := cp.Stop()
		if err != nil {
			t.Fatal(err)
		}
	}()
	cli := cp.NavigatorCtl()
	stdout, stderr, err := cli.Run("get", "pilots,cassandra,elasticsearch")
	t.Log("stdout2", stdout)
	t.Log("stderr2", stderr)
	require.NoError(t, err)
}
