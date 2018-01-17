package validation_test

import (
	"strconv"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/validation"
)

var (
	validNodePoolName         = "valid-name"
	validNodePoolReplicas     = int64(5)
	validNodePoolRoles        = []navigator.ElasticsearchClusterRole{navigator.ElasticsearchRoleData}
	validNodePoolNodeSelector = map[string]string{
		"some": "selector",
	}
	// TODO: expand test cases here
	validNodePoolResources         = &corev1.ResourceRequirements{}
	validNodePoolPersistenceConfig = navigator.ElasticsearchClusterPersistenceConfig{
		Enabled: true,
		Size:    resource.MustParse("10Gi"),
	}

	validImageTag        = "latest"
	validImageRepo       = "something"
	validImagePullPolicy = corev1.PullIfNotPresent
	validImageSpec       = navigator.ImageSpec{
		Tag:        validImageTag,
		Repository: validImageRepo,
		PullPolicy: validImagePullPolicy,
	}

	validSpecPluginsList = []string{"anything"}
	validSpecESImage     = navigator.ElasticsearchImage{
		FsGroup:   int64(1000),
		ImageSpec: validImageSpec,
	}
	validSpecPilotImage = navigator.ElasticsearchPilotImage{
		ImageSpec: validImageSpec,
	}
)

func newValidNodePool(name string, replicas int64, roles ...navigator.ElasticsearchClusterRole) navigator.ElasticsearchClusterNodePool {
	return navigator.ElasticsearchClusterNodePool{
		Name:         name,
		Replicas:     replicas,
		Roles:        roles,
		NodeSelector: validNodePoolNodeSelector,
		Resources:    validNodePoolResources,
		Persistence:  validNodePoolPersistenceConfig,
	}
}

func TestValidateImageSpec(t *testing.T) {
	errorCases := map[string]navigator.ImageSpec{
		"empty image spec": {},
		// missing repo
		"missing repository": {
			Tag:        validImageTag,
			PullPolicy: validImagePullPolicy,
		},
		"missing tag": {
			Repository: validImageRepo,
			PullPolicy: validImagePullPolicy,
		},
		"invalid pullPolicy": {
			Repository: validImageRepo,
			Tag:        validImageTag,
			PullPolicy: "invalid",
		},
	}
	successCases := []navigator.ImageSpec{
		validImageSpec,
		{
			Repository: validImageRepo,
			Tag:        validImageTag,
			PullPolicy: corev1.PullNever,
		},
		{
			Repository: validImageRepo,
			Tag:        validImageTag,
			PullPolicy: corev1.PullIfNotPresent,
		},
		{
			Repository: validImageRepo,
			Tag:        validImageTag,
			PullPolicy: corev1.PullAlways,
		},
		{
			Repository: validImageRepo,
			Tag:        validImageTag,
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			if errs := validation.ValidateImageSpec(&successCase, field.NewPath("test")); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	for n, test := range errorCases {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidateImageSpec(&test, field.NewPath("test"))
			if len(errs) == 0 {
				t.Errorf("Expected errors to be returned for spec %+v but got none", test)
			}
			for _, err := range errs {
				field := err.Field
				if !strings.HasPrefix(field, "test.") &&
					field != "test.tag" &&
					field != "test.repository" &&
					field != "test.pullPolicy" {
					t.Errorf("%s: missing prefix for: %v", n, err)
				}
			}
		})
	}
}

func TestValidateElasticsearchClusterRole(t *testing.T) {
	errorCases := map[string]navigator.ElasticsearchClusterRole{
		"invalid role": navigator.ElasticsearchClusterRole("invalid"),
	}
	successCases := []navigator.ElasticsearchClusterRole{
		navigator.ElasticsearchRoleData,
		navigator.ElasticsearchRoleIngest,
		navigator.ElasticsearchRoleMaster,
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			if errs := validation.ValidateElasticsearchClusterRole(successCase, field.NewPath("test")); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	for n, test := range errorCases {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidateElasticsearchClusterRole(test, field.NewPath("test"))
			if len(errs) == 0 {
				t.Errorf("Expected errors to be returned for spec %+v but got none", test)
			}
			for _, err := range errs {
				field := err.Field
				if !strings.HasPrefix(field, "test") {
					t.Errorf("%s: missing prefix for: %v", n, err)
				}
			}
		})
	}
}

func TestValidateElasticsearchClusterNodePool(t *testing.T) {
	// Invalid as it is missing size parameter
	invalidPersistenceConfig := navigator.ElasticsearchClusterPersistenceConfig{
		Enabled: true,
	}

	invalidRoles := []navigator.ElasticsearchClusterRole{navigator.ElasticsearchClusterRole("invalid")}

	errorCases := map[string]navigator.ElasticsearchClusterNodePool{
		"missing name": {
			Replicas:     validNodePoolReplicas,
			Roles:        validNodePoolRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		"name contains caps": {
			Name:         "Something",
			Replicas:     validNodePoolReplicas,
			Roles:        validNodePoolRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		"name contains symbols": {
			Name:         "something@",
			Replicas:     validNodePoolReplicas,
			Roles:        validNodePoolRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		"negative replicas": {
			Name:         validNodePoolName,
			Replicas:     int64(-1),
			Roles:        validNodePoolRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		"invalid roles": {
			Name:         validNodePoolName,
			Replicas:     validNodePoolReplicas,
			Roles:        invalidRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		"missing roles": {
			Name:         validNodePoolName,
			Replicas:     validNodePoolReplicas,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		"invalid persistence config": {
			Name:         validNodePoolName,
			Replicas:     validNodePoolReplicas,
			Roles:        validNodePoolRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  invalidPersistenceConfig,
		},
	}

	successCases := []navigator.ElasticsearchClusterNodePool{
		{
			Name:         validNodePoolName,
			Replicas:     validNodePoolReplicas,
			Roles:        validNodePoolRoles,
			NodeSelector: validNodePoolNodeSelector,
			Resources:    validNodePoolResources,
			Persistence:  validNodePoolPersistenceConfig,
		},
		{
			Name:  validNodePoolName,
			Roles: validNodePoolRoles,
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			if errs := validation.ValidateElasticsearchClusterNodePool(&successCase, field.NewPath("test")); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	for n, test := range errorCases {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidateElasticsearchClusterNodePool(&test, field.NewPath("test"))
			if len(errs) == 0 {
				t.Errorf("Expected errors to be returned for spec %+v but got none", test)
			}
			for _, err := range errs {
				field := err.Field
				if !strings.HasPrefix(field, "test") &&
					field != "test.name" &&
					field != "test.replicas" &&
					field != "test.roles" &&
					field != "test.nodeSelector" &&
					field != "test.resources" &&
					field != "test.persistence" {
					t.Errorf("%s: missing prefix for: %v", n, err)
				}
			}
		})
	}
}

func TestValidateElasticsearchPersistence(t *testing.T) {
	validSize := resource.MustParse("10Gi")
	errorCases := map[string]navigator.ElasticsearchClusterPersistenceConfig{
		"enabled but no size specified": navigator.ElasticsearchClusterPersistenceConfig{
			Enabled: true,
		},
		"enabled but invalid size specified": navigator.ElasticsearchClusterPersistenceConfig{
			Enabled: true,
			Size:    *resource.NewQuantity(-1, resource.BinarySI),
		},
		"disabled but invalid size specified": navigator.ElasticsearchClusterPersistenceConfig{
			Enabled: false,
			Size:    *resource.NewQuantity(-1, resource.BinarySI),
		},
		"enabled but zero value for size specified": navigator.ElasticsearchClusterPersistenceConfig{
			Enabled: true,
			Size:    *resource.NewQuantity(0, resource.BinarySI),
		},
	}
	successCases := []navigator.ElasticsearchClusterPersistenceConfig{
		{},
		// disabled, but valid size entered
		{
			Enabled: false,
			Size:    validSize,
		},
		{
			Enabled:      false,
			StorageClass: "something",
		},
		{
			Enabled: true,
			Size:    validSize,
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			if errs := validation.ValidateElasticsearchPersistence(&successCase, field.NewPath("test")); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	for n, test := range errorCases {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidateElasticsearchPersistence(&test, field.NewPath("test"))
			if len(errs) == 0 {
				t.Errorf("Expected errors to be returned for spec %+v but got none", test)
			}
			for _, err := range errs {
				field := err.Field
				if !strings.HasPrefix(field, "test") &&
					field != "test.enabled" &&
					field != "test.size" &&
					field != "test.storageClass" {
					t.Errorf("%s: missing prefix for: %v", n, err)
				}
			}
		})
	}
}
func TestValidateElasticsearchClusterSpec(t *testing.T) {
	errorCases := map[string]navigator.ElasticsearchClusterSpec{
		"empty spec": {},
		"valid node pools with duplicate names": {
			Plugins: validSpecPluginsList,
			NodePools: []navigator.ElasticsearchClusterNodePool{
				newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster),
				newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster),
			},
			Pilot:          validSpecPilotImage,
			Image:          validSpecESImage,
			MinimumMasters: 4,
		},
		"missing pilot image": {
			Plugins:        validSpecPluginsList,
			NodePools:      []navigator.ElasticsearchClusterNodePool{newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster)},
			Image:          validSpecESImage,
			MinimumMasters: 2,
		},
		"missing elasticsearch image": {
			Plugins:        validSpecPluginsList,
			NodePools:      []navigator.ElasticsearchClusterNodePool{newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster)},
			Pilot:          validSpecPilotImage,
			MinimumMasters: 2,
		},
		"minimum masters set to low": {
			Plugins:        validSpecPluginsList,
			NodePools:      []navigator.ElasticsearchClusterNodePool{newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster)},
			Pilot:          validSpecPilotImage,
			Image:          validSpecESImage,
			MinimumMasters: 1,
		},
		"minimum masters greater than total number of masters": {
			Plugins:        validSpecPluginsList,
			NodePools:      []navigator.ElasticsearchClusterNodePool{newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster)},
			Pilot:          validSpecPilotImage,
			Image:          validSpecESImage,
			MinimumMasters: 5,
		},
	}
	successCases := []navigator.ElasticsearchClusterSpec{
		{
			NodePools:      []navigator.ElasticsearchClusterNodePool{newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster)},
			Pilot:          validSpecPilotImage,
			Image:          validSpecESImage,
			MinimumMasters: 2,
		},
		{
			Plugins:        validSpecPluginsList,
			NodePools:      []navigator.ElasticsearchClusterNodePool{newValidNodePool("test", 3, navigator.ElasticsearchRoleMaster)},
			Pilot:          validSpecPilotImage,
			Image:          validSpecESImage,
			MinimumMasters: 2,
		},
	}
	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			if errs := validation.ValidateElasticsearchClusterSpec(&successCase, field.NewPath("test")); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	for n, test := range errorCases {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidateElasticsearchClusterSpec(&test, field.NewPath("test"))
			if len(errs) == 0 {
				t.Errorf("Expected errors to be returned for spec %+v but got none", test)
			}
			for _, err := range errs {
				field := err.Field
				if !strings.HasPrefix(field, "test") &&
					field != "test.plugins" &&
					field != "test.nodePools" &&
					field != "test.pilot" &&
					field != "test.image" &&
					field != "test.sysctl" {
					t.Errorf("%s: missing prefix for: %v", n, err)
				}
			}
		})
	}
}
