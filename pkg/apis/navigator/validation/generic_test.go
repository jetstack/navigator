package validation_test

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/jetstack/navigator/pkg/apis/navigator"
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
		Enabled: true,
		Size:    resource.MustParse("10Gi"),
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
	persistenceErrorCases = map[string]navigator.PersistenceConfig{
		"persistence disabled": {
			Enabled: false,
			Size:    resource.MustParse("10Gi"),
		},
		"persistence increased size": {
			Enabled: true,
			Size:    resource.MustParse("25Gi"),
		},
	}
	navigatorClusterConfigErrorCases = map[string]navigator.NavigatorClusterConfig{
		"missing pilot image": {},
	}
)

func init() {
	for title, ec := range imageErrorCases {
		o := validNavigatorClusterConfig.DeepCopy()
		o.PilotImage = ec
		navigatorClusterConfigErrorCases[fmt.Sprintf("pilotimage-%s", title)] = *o
	}
}
