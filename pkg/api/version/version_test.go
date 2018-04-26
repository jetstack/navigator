package version_test

import (
	"testing"

	"github.com/jetstack/navigator/pkg/api/version"
)

func TestUnmarshalJSON(t *testing.T) {
	type testT struct {
		s         string
		v         *version.Version
		expectErr bool
	}
	tests := map[string]testT{
		"unquoted": {
			s:         `3.9.0`,
			expectErr: true,
		},
		"non-integer": {
			s:         `"0.0.x"`,
			expectErr: true,
		},
		"cassandra partial invalid semver with labels": {
			s:         `"X.Y-foo+bar"`,
			expectErr: true,
		},
		"invalid semver with labels": {
			s:         `"X.Y.0-"`,
			expectErr: true,
		},
		// Cassandra versions always include a minor version but Hashicorp
		// go-version (which we currently use for parsing) doesn't require it.
		"partial semver": {
			s: `"3"`,
			v: version.New("3.0.0"),
		},
		"cassandra partial semver": {
			s: `"3.9"`,
			v: version.New("3.9.0"),
		},
		"cassandra partial semver with labels": {
			s: `"3.9-alpha1+dev2"`,
			v: version.New("3.9.0-alpha1+dev2"),
		},
		"valid semver": {
			s: `"3.9.0"`,
			v: version.New("3.9.0"),
		},
		"valid semver with labels": {
			s: `"3.9.0-beta1+dev1"`,
			v: version.New("3.9.0-beta1+dev1"),
		},
	}
	for title, test := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				v := &version.Version{}
				err := v.UnmarshalJSON([]byte(test.s))
				if err == nil {
					if test.expectErr {
						t.Error("Expected an error from Unmarshal but got nil")
					}
					if !test.v.Equal(v) {
						t.Errorf(
							"Version mismatch: expected %s != actual %s",
							test.v, v,
						)
					}
				} else {
					if !test.expectErr {
						t.Errorf("Unexpected error: %s", err)
					} else {
						t.Log(err)
					}
				}
				if err != nil {
					return
				}
				out, err := v.MarshalJSON()
				if err != nil {
					t.Error(err)
				}
				outString := string(out)
				if outString != test.s {
					t.Errorf(
						"JSON marshalling round trip mismatch: %s != %s",
						test.s, outString,
					)
				}
			},
		)
	}
}

func TestDeepCopy(t *testing.T) {
	t.Run(
		"zero value",
		func(t *testing.T) {
			t.Log(version.Version{}.DeepCopy())
		},
	)
	t.Run(
		"validated version",
		func(t *testing.T) {
			t.Log(version.New("3.11.2").DeepCopy())
		},
	)
}
