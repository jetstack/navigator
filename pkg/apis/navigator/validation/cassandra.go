package validation

import (
	"reflect"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
)

func ValidateCassandraClusterNodePool(np *navigator.CassandraClusterNodePool, fldPath *field.Path) field.ErrorList {
	return field.ErrorList{}
}

func ValidateCassandraCluster(c *navigator.CassandraCluster) field.ErrorList {
	allErrs := ValidateObjectMeta(&c.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateCassandraClusterSpec(&c.Spec, field.NewPath("spec"))...)
	return allErrs
}

func ValidateCassandraClusterUpdate(old, new *navigator.CassandraCluster) field.ErrorList {
	allErrs := ValidateCassandraCluster(new)

	fldPath := field.NewPath("spec")

	npPath := fldPath.Child("nodePools")
	for i, newNp := range new.Spec.NodePools {
		idxPath := npPath.Index(i)

		for _, oldNp := range old.Spec.NodePools {
			if newNp.Name == oldNp.Name {
				if newNp.Rack != oldNp.Rack {
					allErrs = append(allErrs, field.Forbidden(idxPath.Child("rack"), "cannot modify rack"))
				}
				if newNp.Datacenter != oldNp.Datacenter {
					allErrs = append(allErrs, field.Forbidden(idxPath.Child("datacenter"), "cannot modify datacenter"))
				}
				if !reflect.DeepEqual(newNp.NodeSelector, oldNp.NodeSelector) {
					allErrs = append(allErrs, field.Forbidden(idxPath.Child("nodeSelector"), "cannot modify nodeSelector"))
				}
				break
			}
		}
	}
	return allErrs
}

func ValidateCassandraClusterSpec(spec *navigator.CassandraClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := ValidateNavigatorClusterConfig(&spec.NavigatorClusterConfig, fldPath)
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
	if spec.Image != nil {
		allErrs = append(
			allErrs,
			ValidateImageSpec(spec.Image, fldPath.Child("image"))...,
		)
	}
	if spec.Version.Equal(emptySemver) {
		allErrs = append(
			allErrs,
			field.Required(fldPath.Child("version"), "must be a semver version"),
		)
	}
	return allErrs
}
