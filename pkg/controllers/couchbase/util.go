package couchbase

import (
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func managedOwnerRef(meta metav1.ObjectMeta) *metav1.OwnerReference {
	for _, ref := range meta.OwnerReferences {
		if ref.APIVersion == navigator.GroupName+"/v1alpha1" && ref.Kind == kindName {
			return &ref
		}
	}
	return nil
}
