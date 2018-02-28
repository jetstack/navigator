package validation

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/coreos/go-semver/semver"

	"github.com/jetstack/navigator/pkg/apis/navigator"
)

var supportedPullPolicies = []string{
	string(corev1.PullNever),
	string(corev1.PullIfNotPresent),
	string(corev1.PullAlways),
	"",
}

var emptySemver = semver.Version{}

func ValidateImageSpec(img *navigator.ImageSpec, fldPath *field.Path) field.ErrorList {
	el := field.ErrorList{}
	if img.Tag == "" {
		el = append(el, field.Required(fldPath.Child("tag"), ""))
	}
	if img.Repository == "" {
		el = append(el, field.Required(fldPath.Child("repository"), ""))
	}
	if img.PullPolicy != corev1.PullNever &&
		img.PullPolicy != corev1.PullIfNotPresent &&
		img.PullPolicy != corev1.PullAlways &&
		img.PullPolicy != "" {
		el = append(el, field.NotSupported(fldPath.Child("pullPolicy"), img.PullPolicy, supportedPullPolicies))
	}
	return el
}

func ValidateNavigatorClusterConfig(cfg *navigator.NavigatorClusterConfig, fldPath *field.Path) field.ErrorList {
	allErrs := ValidateImageSpec(&cfg.PilotImage, fldPath.Child("pilotImage"))
	allErrs = append(allErrs, ValidateNavigatorSecurityContext(&cfg.SecurityContext, fldPath.Child("securityContext"))...)
	sysctlsPath := fldPath.Child("sysctls")
	for i, val := range cfg.Sysctls {
		parts := strings.Split(val, "=")
		if len(parts) != 2 {
			allErrs = append(allErrs, field.Invalid(sysctlsPath.Index(i), val, "should be in format vm.max_map_count=65536"))
		}
	}
	return allErrs
}

func ValidateNavigatorSecurityContext(ctx *navigator.NavigatorSecurityContext, fldPath *field.Path) field.ErrorList {
	el := field.ErrorList{}
	if ctx.RunAsUser != nil {
		if *ctx.RunAsUser < 0 {
			el = append(el, field.Invalid(fldPath.Child("runAsUser"), *ctx.RunAsUser, "must be non-negative"))
		}
	}
	return el
}
