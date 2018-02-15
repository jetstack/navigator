package util

import (
	"fmt"
	"testing"
)

func TestContains(t *testing.T) {
	type testT struct {
		list     interface{}
		elem     interface{}
		contains bool
	}
	tests := []testT{
		{
			list:     []int{1},
			elem:     1,
			contains: true,
		},
		{
			list:     []int{1, 2, 3},
			elem:     1,
			contains: true,
		},
		{
			list:     []int{2},
			elem:     1,
			contains: false,
		},
		{
			list:     []int{},
			elem:     1,
			contains: false,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			actual := Contains(test.list, test.elem)
			if test.contains != actual {
				t.Errorf("expected %t but got %t", test.contains, actual)
			}
		})
	}
}

func TestContainsAll(t *testing.T) {
	type testT struct {
		list     interface{}
		elems    interface{}
		contains bool
	}
	tests := []testT{
		{
			list:     []int{1, 2, 3},
			elems:    []int{1},
			contains: true,
		},
		{
			list:     []int{1, 2, 3},
			elems:    []int{1, 2, 3},
			contains: true,
		},
		{
			list:     []int{1, 2},
			elems:    []int{1, 2, 3},
			contains: false,
		},
		{
			list:     []int{},
			elems:    []int{1},
			contains: false,
		},
		{
			list:     []int{2},
			elems:    []int{1},
			contains: false,
		},
		{
			list:     []int{3, 2},
			elems:    []int{2, 3},
			contains: true,
		},
		{
			list:     []*struct{ int }{ptr(struct{ int }{2}), ptr(struct{ int }{3})},
			elems:    []*struct{ int }{ptr(struct{ int }{3}), ptr(struct{ int }{2})},
			contains: true,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			actual := ContainsAll(test.list, test.elems)
			if test.contains != actual {
				t.Errorf("expected %t but got %t", test.contains, actual)
			}
		})
	}
}

func TestFuzzyEqualLists(t *testing.T) {
	type testT struct {
		left  interface{}
		right interface{}
		equal bool
	}
	tests := []testT{
		{
			left:  []int{1, 2, 3},
			right: []int{1},
			equal: false,
		},
		{
			left:  []int{1, 2, 3},
			right: []int{1, 2, 3},
			equal: true,
		},
		{
			left:  []int{1, 2},
			right: []int{1, 2, 3},
			equal: false,
		},
		{
			left:  []int{},
			right: []int{1},
			equal: false,
		},
		{
			left:  []int{2},
			right: []int{1},
			equal: false,
		},
		{
			left:  []int{3, 2},
			right: []int{2, 3},
			equal: true,
		},
		{
			left:  []*struct{ int }{ptr(struct{ int }{2}), ptr(struct{ int }{3})},
			right: []*struct{ int }{ptr(struct{ int }{3}), ptr(struct{ int }{2})},
			equal: true,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			actual := FuzzyEqualLists(test.left, test.right)
			if test.equal != actual {
				t.Errorf("expected %t but got %t", test.equal, actual)
			}
		})
	}
}

func ptr(o struct{ int }) *struct{ int } {
	return &o
}
