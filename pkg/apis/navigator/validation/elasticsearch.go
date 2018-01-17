package validation

import (
	corev1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
)

var supportedPullPolicies = []string{
	string(corev1.PullNever),
	string(corev1.PullIfNotPresent),
	string(corev1.PullAlways),
	"",
}

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

var supportedESClusterRoles = []string{
	string(navigator.ElasticsearchRoleData),
	string(navigator.ElasticsearchRoleIngest),
	string(navigator.ElasticsearchRoleMaster),
}

func ValidateElasticsearchClusterRole(r navigator.ElasticsearchClusterRole, fldPath *field.Path) field.ErrorList {
	el := field.ErrorList{}
	switch r {
	case navigator.ElasticsearchRoleData:
	case navigator.ElasticsearchRoleIngest:
	case navigator.ElasticsearchRoleMaster:
	default:
		el = append(el, field.NotSupported(fldPath, r, supportedESClusterRoles))
	}
	return el
}

func ValidateElasticsearchClusterNodePool(np *navigator.ElasticsearchClusterNodePool, fldPath *field.Path) field.ErrorList {
	el := ValidateDNS1123Subdomain(np.Name, fldPath.Child("name"))
	el = append(el, ValidateElasticsearchPersistence(&np.Persistence, fldPath.Child("persistence"))...)
	rolesPath := fldPath.Child("roles")
	if len(np.Roles) == 0 {
		el = append(el, field.Required(rolesPath, "at least one role must be specified"))
	}
	for i, r := range np.Roles {
		idxPath := rolesPath.Index(i)
		el = append(el, ValidateElasticsearchClusterRole(r, idxPath)...)
	}
	if np.Replicas < 0 {
		el = append(el, field.Invalid(fldPath.Child("replicas"), np.Replicas, "must be greater than zero"))
	}
	// TODO: call k8s.io/kubernetes/pkg/apis/core/validation.ValidateResourceRequirements on np.Resources
	// this will require vendoring kubernetes/kubernetes and switching to use the corev1 ResourceRequirements
	// struct
	return el
}

func ValidateElasticsearchPersistence(cfg *navigator.ElasticsearchClusterPersistenceConfig, fldPath *field.Path) field.ErrorList {
	el := field.ErrorList{}
	if cfg.Enabled && len(cfg.Size) == 0 {
		el = append(el, field.Required(fldPath.Child("size"), ""))
	}
	// TODO: validate size quantity
	return el
}

func ValidateElasticsearchClusterSpec(spec *navigator.ElasticsearchClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := ValidateImageSpec(&spec.Image.ImageSpec, fldPath.Child("image"))
	allErrs = append(allErrs, ValidateImageSpec(&spec.Pilot.ImageSpec, fldPath.Child("pilot"))...)
	npPath := fldPath.Child("nodePools")
	allNames := sets.String{}
	for i, np := range spec.NodePools {
		idxPath := npPath.Index(i)
		if allNames.Has(np.Name) {
			allErrs = append(allErrs, field.Duplicate(idxPath.Child("name"), np.Name))
		} else {
			allNames.Insert(np.Name)
		}
		allErrs = append(allErrs, ValidateElasticsearchClusterNodePool(&np, idxPath)...)
	}
	return allErrs
}

func ValidateElasticsearchCluster(esc *navigator.ElasticsearchCluster) field.ErrorList {
	allErrs := ValidateObjectMeta(&esc.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateElasticsearchClusterSpec(&esc.Spec, field.NewPath("spec"))...)
	return allErrs
}
