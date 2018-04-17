package api

import (
	"fmt"
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/util/ptr"
)

func TestCountElasticsearchMasters(t *testing.T) {
	type testT struct {
		in    []v1alpha1.ElasticsearchClusterNodePool
		count int32
	}
	tests := []testT{
		{
			in: []v1alpha1.ElasticsearchClusterNodePool{
				{
					Replicas: ptr.Int32(2),
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
			},
			count: 2,
		},
		{
			in: []v1alpha1.ElasticsearchClusterNodePool{
				{
					Replicas: ptr.Int32(2),
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
				{
					Replicas: ptr.Int32(2),
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
			},
			count: 4,
		},
		{
			in: []v1alpha1.ElasticsearchClusterNodePool{
				{
					Replicas: ptr.Int32(2),
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleMaster},
				},
				{
					Replicas: ptr.Int32(2),
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleIngest},
				},
				{
					Replicas: ptr.Int32(2),
					Roles:    []v1alpha1.ElasticsearchClusterRole{v1alpha1.ElasticsearchRoleData},
				},
			},
			count: 2,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := CountElasticsearchMasters(test.in)
			if actual != test.count {
				t.Errorf("expected %d but got %d", test.count, actual)
			}
		})
	}
}

func TestContainsElasticsearchRole(t *testing.T) {
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
			actual := ContainsElasticsearchRole(test.set, test.lookup)
			if actual != test.has {
				t.Errorf("expected %t but got %t", test.has, actual)
			}
		})
	}
}
