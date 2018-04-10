package validation_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/coreos/go-semver/semver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/validation"
	"github.com/jetstack/navigator/pkg/util/ptr"
)

var (
	validSemver          = *semver.New("5.6.2")
	validImageTag        = "latest"
	validImageRepo       = "something"
	validImagePullPolicy = corev1.PullIfNotPresent
	validImageSpec       = navigator.ImageSpec{
		Tag:        validImageTag,
		Repository: validImageRepo,
		PullPolicy: validImagePullPolicy,
	}
	validNodePoolPersistenceConfig = navigator.PersistenceConfig{
		Size: resource.MustParse("10Gi"),
	}
	validNavigatorClusterConfig = navigator.NavigatorClusterConfig{
		PilotImage: validImageSpec,
	}
	imageErrorCases = map[string]navigator.ImageSpec{
		"missing repository": {
			Tag:        validImageTag,
			PullPolicy: validImagePullPolicy,
		},
		"missing tag": {
			Repository: validImageRepo,
			PullPolicy: validImagePullPolicy,
		},
	}
	persistenceErrorCases = map[string]*navigator.PersistenceConfig{
		"persistence disabled": nil,
		"persistence increased size": {
			Size: resource.MustParse("25Gi"),
		},
	}
	navigatorClusterConfigErrorCases = map[string]navigator.NavigatorClusterConfig{
		"missing pilot image": {},
	}
)

func TestValidatePersistenceConfig(t *testing.T) {
	validSize := resource.MustParse("10Gi")
	errorCases := map[string]navigator.PersistenceConfig{
		"no size specified": navigator.PersistenceConfig{},
		"invalid size specified": navigator.PersistenceConfig{
			Size: *resource.NewQuantity(-1, resource.BinarySI),
		},
	}
	successCases := map[string]navigator.PersistenceConfig{
		"valid size and no storage class set": {
			Size: validSize,
		},
		"storage class and size set": {
			Size:         validSize,
			StorageClass: ptr.String("something"),
		},
	}

	for n, successCase := range successCases {
		t.Run(n, func(t *testing.T) {
			if errs := validation.ValidatePersistenceConfig(&successCase, field.NewPath("test")); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	for n, test := range errorCases {
		t.Run(n, func(t *testing.T) {
			errs := validation.ValidatePersistenceConfig(&test, field.NewPath("test"))
			if len(errs) == 0 {
				t.Errorf("Expected errors to be returned for spec %+v but got none", test)
			}
			for _, err := range errs {
				field := err.Field
				if !strings.HasPrefix(field, "test") &&
					field != "test.size" &&
					field != "test.storageClass" {
					t.Errorf("%s: missing prefix for: %v", n, err)
				}
			}
		})
	}
}

func init() {
	for title, ec := range imageErrorCases {
		o := validNavigatorClusterConfig.DeepCopy()
		o.PilotImage = ec
		navigatorClusterConfigErrorCases[fmt.Sprintf("pilotimage-%s", title)] = *o
	}
}
