package configmap

import (
	"fmt"
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func TestCountMasterReplicas(t *testing.T) {
	type testT struct {
		in    []v1alpha1.ElasticsearchClusterNodePool
		count int64
	}
	tests := []testT{
		{
			in: []v1alpha1.ElasticsearchClusterNodePool{
				{
					Replicas: 2,
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
			},
			count: 2,
		},
		{
			in: []v1alpha1.ElasticsearchClusterNodePool{
				{
					Replicas: 2,
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
				{
					Replicas: 2,
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
			},
			count: 4,
		},
		{
			in: []v1alpha1.ElasticsearchClusterNodePool{
				{
					Replicas: 2,
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
				{
					Replicas: 2,
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleIngest},
				},
				{
					Replicas: 2,
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleData},
				},
			},
			count: 2,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := countMasterReplicas(test.in)
			if actual != test.count {
				t.Errorf("expected %d but got %d", test.count, actual)
			}
		})
	}
}

func TestHasRole(t *testing.T) {
	type testT struct {
		set    []v1alpha1.ElasticsearchClusterRole
		lookup v1alpha1.ElasticsearchClusterRole
		has    bool
	}
	tests := []testT{
		{
			set:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
			lookup: v1alpha1.ElasticsearchRoleMaster,
			has:    true,
		},
		{
			set:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleIngest},
			lookup: v1alpha1.ElasticsearchRoleMaster,
			has:    false,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := hasRole(test.set, test.lookup)
			if actual != test.has {
				t.Errorf("expected %t but got %t", test.has, actual)
			}
		})
	}
}

func TestCalculateQuorom(t *testing.T) {
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
			actual := calculateQuorom(test.in)
			if actual != test.out {
				t.Errorf("expected %d but got %d", test.out, actual)
			}
		})
	}
}
