package v5

import (
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func TestParseHealth(t *testing.T) {
	type testT struct {
		str      string
		expected v1alpha1.ElasticsearchClusterHealth
	}
	tests := []testT{
		{
			str:      "grEeN",
			expected: v1alpha1.ElasticsearchClusterHealthGreen,
		},
		{
			str:      "reD",
			expected: v1alpha1.ElasticsearchClusterHealthRed,
		},
		{
			str:      "YELLOW",
			expected: v1alpha1.ElasticsearchClusterHealthYellow,
		},
		{
			str:      "abcdefgh",
			expected: "abcdefgh",
		},
		{
			str:      "ABCdef",
			expected: "ABCdef",
		},
	}
	testFn := func(test testT) func(*testing.T) {
		return func(t *testing.T) {
			actual := parseHealth(test.str)
			if actual != test.expected {
				t.Errorf("Expected health status to equal %s but got %s", test.expected, actual)
			}
		}
	}
	for _, test := range tests {
		t.Run("", testFn(test))
	}
}
