package v3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	shutil "github.com/termie/go-shutil"

	"github.com/jetstack/navigator/internal/test/util/testfs"
	"github.com/jetstack/navigator/pkg/api/version"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/cassandra/nodetool"
	fakenodetool "github.com/jetstack/navigator/pkg/cassandra/nodetool/fake"
	"github.com/jetstack/navigator/pkg/config"
)

func TestPilotSyncFunc(t *testing.T) {
	type testT struct {
		pilot      *v1alpha1.Pilot
		nodetool   nodetool.Interface
		assertions func(t *testing.T, pilotIn, pilotOut *v1alpha1.Pilot)
		expectErr  bool
	}
	tests := map[string]testT{
		"pilot found with nil version": {
			pilot:    &v1alpha1.Pilot{},
			nodetool: fakenodetool.New().SetVersion("3.11.1"),
			assertions: func(t *testing.T, pilotIn, pilotOut *v1alpha1.Pilot) {
				expectedVersion := version.New("3.11.1")
				actualVersion := pilotOut.Status.Cassandra.Version
				if !expectedVersion.Equal(actualVersion) {
					t.Errorf("Version mismatch. Expected %s. Got %s.", expectedVersion, actualVersion)
				}
			},
		},
		"pilot found with different version": {
			pilot: &v1alpha1.Pilot{
				Status: v1alpha1.PilotStatus{
					Cassandra: &v1alpha1.CassandraPilotStatus{
						Version: version.New("3.0.0"),
					},
				},
			},
			nodetool: fakenodetool.New().SetVersion("3.11.1"),
			assertions: func(t *testing.T, pilotIn, pilotOut *v1alpha1.Pilot) {
				expectedVersion := version.New("3.11.1")
				actualVersion := pilotOut.Status.Cassandra.Version
				if !expectedVersion.Equal(actualVersion) {
					t.Errorf("Version mismatch. Expected %s. Got %s.", expectedVersion, actualVersion)
				}
			},
		},
		"nodetool version failure causes nil status version": {
			pilot: &v1alpha1.Pilot{
				Status: v1alpha1.PilotStatus{
					Cassandra: &v1alpha1.CassandraPilotStatus{
						Version: version.New("3.0.0"),
					},
				},
			},
			nodetool: fakenodetool.New().SetVersionError("simulated nodetool error"),
			assertions: func(t *testing.T, pilotIn, pilotOut *v1alpha1.Pilot) {
				actualVersion := pilotOut.Status.Cassandra.Version
				if actualVersion != nil {
					t.Errorf("Expected nil version. Got: %s", actualVersion)
				}
			},
		},
	}
	for title, test := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				p := &Pilot{
					nodeTool: test.nodetool,
				}
				pilot := test.pilot.DeepCopy()
				err := p.syncFunc(pilot)
				if err != nil {
					if !test.expectErr {
						t.Errorf("Unexpected error: %s", err)
					}
				} else {
					if test.expectErr {
						t.Error("Missing error")
					}
				}
				if test.assertions != nil {
					test.assertions(t, test.pilot, pilot)
				}
			},
		)
	}
}

func TestWriteConfig(t *testing.T) {
	t.Run(
		"create cassandra.yaml",
		func(t *testing.T) {
			tfs := testfs.New(t)
			etc := tfs.TempDir("etc")
			cassConfigPath := etc + "/cassandra"
			err := shutil.CopyTree("testdata", cassConfigPath, nil)
			require.NoError(t, err)

			pilotResource := &v1alpha1.Pilot{}
			p := &Pilot{
				Options: &PilotOptions{
					CassandraConfigPath: cassConfigPath,
				},
			}

			err = p.WriteConfig(pilotResource)
			require.NoError(t, err)
			cassandraYaml := cassConfigPath + "/cassandra.yaml"
			cfg, err := config.NewFromYaml(cassandraYaml)
			require.NoError(t, err)
			assert.Nil(t, cfg.Get("listen_address"))
			assert.Nil(t, cfg.Get("listen_interface"))
			assert.Nil(t, cfg.Get("broadcast_address"))
			assert.Nil(t, cfg.Get("rpc_address"))
			assert.Equal(t, CassSnitch, cfg.Get("endpoint_snitch"))
			seedProviders := cfg.Get("seed_provider").([]interface{})
			assert.Len(t, seedProviders, 1)
			seedProvider := seedProviders[0].(map[interface{}]interface{})
			assert.Equal(t, CassSeedProvider, seedProvider["class_name"])
			parameters := seedProvider["parameters"].([]interface{})
			assert.Len(t, parameters, 1)
			parameter := parameters[0].(map[interface{}]interface{})
			assert.Equal(t, "", parameter["seeds"])
		},
	)
	t.Run(
		"create cassandra-rackdc.properties",
		func(t *testing.T) {
			tfs := testfs.New(t)
			etc := tfs.TempDir("etc")
			cassConfigPath := etc + "/cassandra"
			err := shutil.CopyTree("testdata", cassConfigPath, nil)
			require.NoError(t, err)

			expectedRack := "rack-name-foo"
			expectedDC := "dc-bar"

			pilotResource := &v1alpha1.Pilot{}
			p := &Pilot{
				Options: &PilotOptions{
					CassandraConfigPath: cassConfigPath,
					CassandraRack:       expectedRack,
					CassandraDC:         expectedDC,
				},
			}

			err = p.WriteConfig(pilotResource)
			require.NoError(t, err)

			cfg, err := config.NewFromProperties(cassConfigPath + "/cassandra-rackdc.properties")
			require.NoError(t, err)
			assert.Equal(t, expectedRack, cfg.Get("rack"))
			assert.Equal(t, expectedDC, cfg.Get("dc"))
		},
	)
}
