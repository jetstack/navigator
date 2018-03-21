package v3

import (
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/cassandra/nodetool"
	fakenodetool "github.com/jetstack/navigator/pkg/cassandra/nodetool/fake"
	"github.com/jetstack/navigator/pkg/cassandra/version"
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
