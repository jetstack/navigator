package validation

import (
	"fmt"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/util"
)

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
	// this will require vendoring kubernetes/kubernetes.
	return el
}

func ValidateElasticsearchPersistence(cfg *navigator.PersistenceConfig, fldPath *field.Path) field.ErrorList {
	el := field.ErrorList{}
	if cfg.Enabled && cfg.Size.IsZero() {
		el = append(el, field.Required(fldPath.Child("size"), ""))
	}
	if cfg.Size.Sign() == -1 {
		el = append(el, field.Invalid(fldPath.Child("size"), cfg.Size, "must be greater than zero"))
	}
	return el
}

func ValidateElasticsearchClusterSpec(spec *navigator.ElasticsearchClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := ValidateNavigatorClusterConfig(&spec.NavigatorClusterConfig, fldPath)
	if spec.Image != nil {
		allErrs = append(allErrs, ValidateImageSpec(spec.Image, fldPath.Child("image"))...)
	}
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

	numMasters := countElasticsearchMasters(spec.NodePools)
	quorum := util.CalculateQuorum(numMasters)
	switch {
	case numMasters == 0:
		allErrs = append(allErrs, field.Invalid(npPath, numMasters, "must be at least one master node"))
	case spec.MinimumMasters == 0:
		// do nothing, navigator-controller will automatically calculate &
		// manage the minimumMasters required for the cluster
	case spec.MinimumMasters < quorum:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("minimumMasters"), spec.MinimumMasters, fmt.Sprintf("must be a minimum of %d to avoid a split brain scenario", quorum)))
	case spec.MinimumMasters > numMasters:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("minimumMasters"), spec.MinimumMasters, fmt.Sprintf("cannot be greater than the total number of master nodes")))
	}

	if spec.Version.Equal(emptySemver) {
		allErrs = append(allErrs, field.Required(fldPath.Child("version"), "must be a semver version"))
	}
	return allErrs
}

func ValidateElasticsearchCluster(esc *navigator.ElasticsearchCluster) field.ErrorList {
	allErrs := ValidateObjectMeta(&esc.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateElasticsearchClusterSpec(&esc.Spec, field.NewPath("spec"))...)
	return allErrs
}

func countElasticsearchMasters(pools []navigator.ElasticsearchClusterNodePool) int32 {
	masters := int32(0)
	for _, pool := range pools {
		if containsElasticsearchRole(pool.Roles, navigator.ElasticsearchRoleMaster) {
			masters += pool.Replicas
		}
	}
	return masters
}

func containsElasticsearchRole(set []navigator.ElasticsearchClusterRole, role navigator.ElasticsearchClusterRole) bool {
	for _, s := range set {
		if s == role {
			return true
		}
	}
	return false
}
