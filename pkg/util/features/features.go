package features

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	utilfeature "github.com/jetstack/navigator/pkg/util/feature"
)

const (
	Example = "Example"
)

// InitFeatureGates are the default feature gates for the init command
var InitFeatureGates = FeatureList{
	Example: utilfeature.FeatureSpec{Default: false, PreRelease: utilfeature.Alpha},
}

// FeatureList represents a list of feature gates
type FeatureList map[string]utilfeature.FeatureSpec

// Enabled indicates whether a feature name has been enabled
func Enabled(featureList map[string]bool, featureName string) bool {
	return featureList[string(featureName)]
}

// Supports indicates whether a feature name is supported on the given
// feature set
func Supports(featureList FeatureList, featureName string) bool {
	for k := range featureList {
		if featureName == string(k) {
			return true
		}
	}
	return false
}

// Keys returns a slice of feature names for a given feature set
func Keys(featureList FeatureList) []string {
	var list []string
	for k := range featureList {
		list = append(list, string(k))
	}
	return list
}

// KnownFeatures returns a slice of strings describing the FeatureList features.
func KnownFeatures(f *FeatureList) []string {
	var known []string
	for k, v := range *f {
		pre := ""
		if v.PreRelease != utilfeature.GA {
			pre = fmt.Sprintf("%s - ", v.PreRelease)
		}
		known = append(known, fmt.Sprintf("%s=true|false (%sdefault=%t)", k, pre, v.Default))
	}
	sort.Strings(known)
	return known
}

// NewFeatureGate parses a string of the form "key1=value1,key2=value2,..." into a
// map[string]bool of known keys or returns an error.
func NewFeatureGate(f *FeatureList, value string) (map[string]bool, error) {
	featureGate := map[string]bool{}
	for _, s := range strings.Split(value, ",") {
		if len(s) == 0 {
			continue
		}

		arr := strings.SplitN(s, "=", 2)
		if len(arr) != 2 {
			return nil, fmt.Errorf("missing bool value for feature-gate key:%s", s)
		}

		k := strings.TrimSpace(arr[0])
		v := strings.TrimSpace(arr[1])

		if !Supports(*f, k) {
			return nil, fmt.Errorf("unrecognized feature-gate key: %s", k)
		}

		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value %v for feature-gate key: %s, use true|false instead", v, k)
		}
		featureGate[k] = boolValue
	}

	return featureGate, nil
}
