package validation

import (
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
)

func ValidateCassandraClusterNodePool(np *navigator.CassandraClusterNodePool, fldPath *field.Path) field.ErrorList {
	return field.ErrorList{}
}

func ValidateCassandraCluster(cass *navigator.CassandraCluster) field.ErrorList {
	allErrs := ValidateObjectMeta(&cass.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateCassandraClusterSpec(&cass.Spec, field.NewPath("spec"))...)
	return allErrs
}

func ValidateCassandraClusterUpdate(old, new *navigator.CassandraCluster) field.ErrorList {
	allErrs := ValidateCassandraCluster(new)

	return allErrs
}

func ValidateCassandraClusterSpec(spec *navigator.CassandraClusterSpec, fldPath *field.Path) field.ErrorList {
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
		allErrs = append(allErrs, ValidateCassandraClusterNodePool(&np, idxPath)...)
	}
	return allErrs
}
