package validation

import (
	"github.com/jetstack/navigator/pkg/apis/navigator"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateCassandraCluster(esc *navigator.CassandraCluster) field.ErrorList {
	allErrs := ValidateObjectMeta(&esc.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateCassandraClusterSpec(&esc.Spec, field.NewPath("spec"))...)
	return allErrs
}

func ValidateCassandraClusterSpec(spec *navigator.CassandraClusterSpec, fldPath *field.Path) field.ErrorList {
	allErrs := ValidateNavigatorClusterConfig(&spec.NavigatorClusterConfig, fldPath)
	return allErrs
}
