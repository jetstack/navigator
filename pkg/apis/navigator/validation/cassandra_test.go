package validation_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/validation"
)

var (
	validCassCluster = &navigator.CassandraCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: navigator.CassandraClusterSpec{
			Version: validSemver,
			Image:   &validImageSpec,
			NavigatorClusterConfig: validNavigatorClusterConfig,
		},
	}
)

func TestValidateCassandraCluster(t *testing.T) {
	type testT struct {
		cluster       *navigator.CassandraCluster
		errorExpected bool
	}

	tests := map[string]testT{
		"valid cluster": {
			cluster: validCassCluster,
		},
	}

	setNavigatorClusterConfig := func(
		c *navigator.CassandraCluster,
		ncc navigator.NavigatorClusterConfig,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.NavigatorClusterConfig = ncc
		return c
	}

	for title, ncc := range navigatorClusterConfigErrorCases {
		tests[title] = testT{
			cluster:       setNavigatorClusterConfig(validCassCluster, ncc),
			errorExpected: true,
		}
	}

	setImage := func(
		c *navigator.CassandraCluster,
		image *navigator.ImageSpec,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.Image = image
		return c
	}

	for title, image := range imageErrorCases {
		tests[title] = testT{
			cluster:       setImage(validCassCluster, &image),
			errorExpected: true,
		}
	}

	for title, tc := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				errs := validation.ValidateCassandraCluster(tc.cluster)
				if tc.errorExpected && len(errs) == 0 {
					t.Errorf("expected error but got none")
				}
				if !tc.errorExpected && len(errs) != 0 {
					t.Errorf("unexpected errors: %s", errs)
				}
				for _, e := range errs {
					t.Logf("error string is: %s", e)
				}
			},
		)
	}
}
