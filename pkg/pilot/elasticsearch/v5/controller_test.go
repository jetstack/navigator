package v5

import (
	"reflect"
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func TestParseHealth(t *testing.T) {
	type testT struct {
		str      string
		expected *v1alpha1.ElasticsearchClusterHealth
	}
	greenHealth := v1alpha1.ElasticsearchClusterHealthGreen
	redHealth := v1alpha1.ElasticsearchClusterHealthRed
	yellowHealth := v1alpha1.ElasticsearchClusterHealthYellow
	tests := []testT{
		{
			str:      "grEeN",
			expected: &greenHealth,
		},
		{
			str:      "reD",
			expected: &redHealth,
		},
		{
			str:      "YELLOW",
			expected: &yellowHealth,
		},
		{
			str: "abcdefgh",
		},
		{
			str: "ABCdef",
		},
	}
	testFn := func(test testT) func(*testing.T) {
		return func(t *testing.T) {
			actual := parseHealth(test.str)
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("Expected health status to equal %v but got %v", test.expected, actual)
			}
		}
	}
	for _, test := range tests {
		t.Run("", testFn(test))
	}
}
