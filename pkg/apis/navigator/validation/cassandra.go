package validation

import (
	"fmt"
	"reflect"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/jetstack/navigator/pkg/apis/navigator"
)

func ValidateCassandraClusterNodePool(np *navigator.CassandraClusterNodePool, fldPath *field.Path) field.ErrorList {
	// TODO: call k8s.io/kubernetes/pkg/apis/core/validation.ValidateResourceRequirements on np.Resources
	// this will require vendoring kubernetes/kubernetes.

	allErrs := field.ErrorList{}
	if np.Seeds > np.Replicas {
		allErrs = append(allErrs,
			field.Invalid(
				fldPath.Child("seeds"),
				np.Seeds,
				fmt.Sprintf("number of seeds cannot be greater than number of replicas (%d)", np.Replicas)),
		)
	}

	return allErrs
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
				if !reflect.DeepEqual(newNp.Persistence, oldNp.Persistence) {
					if oldNp.Persistence.Enabled {
						allErrs = append(allErrs, field.Forbidden(idxPath.Child("persistence"), "cannot modify persistence configuration once enabled"))
					}
				}

				restoreReplicas := newNp.Replicas
				newNp.Replicas = oldNp.Replicas

				restorePersistence := newNp.Persistence
				newNp.Persistence = oldNp.Persistence

				if !reflect.DeepEqual(newNp, oldNp) {
					allErrs = append(allErrs, field.Forbidden(field.NewPath("spec"), "updates to nodepool for fields other than 'replicas' and 'persistence' are forbidden."))
				}
				newNp.Replicas = restoreReplicas
				newNp.Persistence = restorePersistence

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
	return allErrs
}
