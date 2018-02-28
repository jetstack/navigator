package validation_test

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	corev1 "k8s.io/api/core/v1"

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
