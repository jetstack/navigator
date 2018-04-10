package version

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-semver/semver"
	utilerror "k8s.io/apimachinery/pkg/util/errors"
)

// Version represents a Cassandra database server version.
// Cassandra does not adhere to semver.
// A Cassandra version string may omit the patch version.
// In Navigator we query the version of the running Cassandra service via the JMX interface.
// It is returned as a string via:
// StorageService.getReleaseVersion: https://github.com/apache/cassandra/blob/cassandra-3.11.2/src/java/org/apache/cassandra/service/StorageService.java#L2863
// FBUtilities.getReleaseVersionString: https://github.com/apache/cassandra/blob/cassandra-3.11.2/src/java/org/apache/cassandra/utils/FBUtilities.java#L326
// Which appears to read the version string from a `Properties` API which appears to be created via this XML file: https://github.com/apache/cassandra/blob/cassandra-3.11.2/build.xml#L790
// Internally, Cassandra converts the version string to a `CassandraVersion` object which supports rich comparison.
// See https://github.com/apache/cassandra/blob/cassandra-3.11.2/src/java/org/apache/cassandra/utils/CassandraVersion.java
// In Navigator we parse the Cassandra version string as early as possible, into a similar Cassandra Version object.
// This also fixes the missing Patch number and stores the version internally as a semver.
// It also keeps a reference to the original version string so that we can report that in our API.
// So that the version reported in our API matches the version that an administrator expects.
type Version struct {
	versionString string
	semver        semver.Version
	Major         int64
	Minor         int64
	Patch         int64
}

func New(s string) *Version {
	v := &Version{}
	err := v.set(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (v *Version) Equal(versionB *Version) bool {
	return v.semver.Equal(versionB.semver)
}

// TODO: Add tests for this
func (v *Version) LessThan(versionB *Version) bool {
	return v.semver.LessThan(versionB.semver)
}

func (v *Version) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return v.set(s)
}

func (v *Version) set(cassVersionString string) error {
	var versionsTried []string
	var errs []error

	err := v.semver.Set(cassVersionString)
	if err != nil {
		versionsTried = append(versionsTried, cassVersionString)
		errs = append(errs, err)

		semverString := maybeAddMissingPatchVersion(cassVersionString)
		if semverString != cassVersionString {
			err = v.semver.Set(semverString)
			versionsTried = append(versionsTried, semverString)
			errs = append(errs, err)
		}
	}
	if err != nil {
		return fmt.Errorf(
			"unable to parse Cassandra version as semver. Versions tried: '%s'. Errors: %s.",
			strings.Join(versionsTried, "','"),
			utilerror.NewAggregate(errs),
		)
	}
	v.versionString = cassVersionString
	v.Major = v.semver.Major
	v.Minor = v.semver.Minor
	v.Patch = v.semver.Patch
	return nil
}

var _ json.Unmarshaler = &Version{}

func maybeAddMissingPatchVersion(v string) string {
	mmpAndLabels := strings.SplitN(v, "-", 2)
	mmp := mmpAndLabels[0]
	mmpParts := strings.SplitN(mmp, ".", 3)
	if len(mmpParts) == 2 {
		mmp = mmp + ".0"
	}
	mmpAndLabels[0] = mmp
	return strings.Join(mmpAndLabels, "-")
}

func (v Version) String() string {
	return v.versionString
}

func (v *Version) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(v.String())), nil
}

var _ json.Marshaler = &Version{}

func (v *Version) Semver() string {
	return v.semver.String()
}

func (v *Version) BumpPatch() *Version {
	sv := semver.New(v.Semver())
	sv.BumpPatch()
	return New(sv.String())
}
func (v *Version) BumpMinor() *Version {
	sv := semver.New(v.Semver())
	sv.BumpMinor()
	return New(sv.String())
}
func (v *Version) BumpMajor() *Version {
	sv := semver.New(v.Semver())
	sv.BumpMajor()
	return New(sv.String())
}
