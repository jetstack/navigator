package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/testing_frameworks/integration"

	"github.com/jetstack/navigator/internal/test/util/testfs"
)

// MakeKubeConfig creates a kube/config file which is suitable for communicating
// with the API server started by ``integration.ControlPlane``.
func MakeKubeConfig(f *integration.ControlPlane, path string) error {
	ctl := f.KubeCtl()
	ctl.Opts = []string{"--kubeconfig", path}

	_, _, err := ctl.Run(
		"config",
		"set-credentials",
		"user1",
	)
	if err != nil {
		return errors.Wrap(err, "unable to create user")
	}

	_, _, err = ctl.Run(
		"config",
		"set-cluster",
		"integration1",
		"--server",
		f.APIURL().String(),
		"--certificate-authority",
		filepath.Join(f.APIServer.CertDir, "apiserver.crt"),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create cluster")
	}

	_, _, err = ctl.Run(
		"config",
		"set-context",
		"default",
		"--cluster", "integration1",
		"--user", "user1",
	)
	if err != nil {
		return errors.Wrap(err, "unable to create context")
	}

	_, _, err = ctl.Run(
		"config",
		"use-context",
		"default",
	)
	if err != nil {
		return errors.Wrap(err, "unable to use context")
	}
	return nil
}

func TestPilotCassandra(t *testing.T) {
	// Start the API server with CustomResourceSubresources feature enabled.
	// Navigator pilot calls on UpdateStatus on Pilot resources for example.
	cp := &integration.ControlPlane{
		APIServer: &integration.APIServer{
			Args: []string{
				"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
				"--cert-dir={{ .CertDir }}",
				"--insecure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
				"--insecure-bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
				"--secure-port=0",
				"--feature-gates=CustomResourceSubresources=true",
				"-v=4", "--alsologtostderr",
			},
			// Out: os.Stdout,
			// Err: os.Stderr,
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
	cli := cp.KubeCtl()

	// Create a CRD for the Navigator Pilot resource type.
	// This allows us to test the Pilot against the Kubernetes API server,
	// without having to start the Navigator API server and configure aggregation etc.
	crdPath, err := filepath.Abs("testdata/pilot_crd.yaml")
	require.NoError(t, err)
	stdout, stderr, err := cli.Run(
		"apply",
		"--filename",
		crdPath,
	)
	t.Log("stdout", stdout)
	t.Log("stderr", stderr)
	require.NoError(t, err)

	// Create a Pilot resource for our Pilot process to watch and update.
	stdout, stderr, err = cli.Run(
		"create",
		"namespace",
		"ns1",
	)
	t.Log("stdout", stdout)
	t.Log("stderr", stderr)
	require.NoError(t, err)
	pilotResourcePath, err := filepath.Abs("testdata/pilot.yaml")
	require.NoError(t, err)
	stdout, stderr, err = cli.Run(
		"apply",
		"--filename",
		pilotResourcePath,
	)
	t.Log("stdout", stdout)
	t.Log("stderr", stderr)
	require.NoError(t, err)

	// Temporary configuration directories.
	tfs := testfs.New(t)
	kubeConfig := tfs.TempPath("kube.config")
	cassConfig := tfs.TempDir("etc_cassandra")
	pilotConfigDir := tfs.TempDir("etc/pilot")
	// A fake cassandra binary for the Pilot to execute.
	cassPath, err := filepath.Abs("testdata/fake_cassandra")
	require.NoError(t, err)
	err = MakeKubeConfig(cp, kubeConfig)
	require.NoError(t, err)

	pilotPath, pilotPathFound := os.LookupEnv("TEST_ASSET_NAVIGATOR_PILOT_CASSANDRA")
	if !pilotPathFound {
		t.Fatal(
			"Please set environment variable TEST_ASSET_NAVIGATOR_PILOT_CASSANDRA " +
				"with the path to the navigator-pilot-cassandra binary under test.")
	}

	expectedClusterName := "cluster1"
	cmd := exec.Command(
		pilotPath,
		"--pilot-name", "pilot1",
		"--pilot-namespace", "ns1",
		"--kubeconfig", kubeConfig,
		"--v=4", "--alsologtostderr",
		"--leader-elect=false",
		"--cassandra-cluster-name", expectedClusterName,
		"--cassandra-config-path", cassConfig,
		"--cassandra-path", cassPath,
		"--cassandra-rack", "rack-for-test",
		"--cassandra-dc", "dc-for-test",
		"--config-dir", pilotConfigDir,
	)
	startDetectStream := gbytes.NewBuffer()
	// The fake Cassandra script echos "FAKE CASSANDRA" in a loop
	ready := startDetectStream.Detect("FAKE CASSANDRA")
	cmd.Stdout = startDetectStream
	cmd.Stderr = os.Stderr
	// This will return when the executable writes "FAKE CASSANDRA" to stdout.
	err = cmd.Start()
	require.NoError(t, err)
	defer func() {
		// XXX: Pilot doesn't respond to SIGKILL
		err := cmd.Process.Signal(os.Interrupt)
		require.NoError(t, err)
		err = cmd.Wait()
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 255 {
					// XXX Pilot should return 0 if it exits cleanly after a signal.
					return
				}
			}
		}
		require.NoError(t, err)
	}()
	select {
	case <-ready:
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for process to start")
	}
	// The pilot has executed the Cassandra sub-process, so it should already
	// have written the configuration files as part of the pre-start hook
	// mechanism.
	// We did not provide an existing cassandra.yaml or cassandra-rackdc.properties.
	// So we expect thos files to now exist.
	assert.FileExists(t, cassConfig+"/cassandra.yaml")
	assert.FileExists(t, cassConfig+"/cassandra-rackdc.properties")
}
