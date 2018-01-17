package util

import (
	"fmt"
	"testing"
)

func TestCalculateQuorum(t *testing.T) {
	type testT struct {
		in  int64
		out int64
	}
	tests := []testT{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{5, 3},
		{6, 4},
		{7, 4},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := CalculateQuorum(test.in)
			if actual != test.out {
				t.Errorf("expected %d but got %d", test.out, actual)
			}
		})
	}
}
