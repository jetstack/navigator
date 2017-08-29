package couchbase

import (
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func managedOwnerRef(meta metav1.ObjectMeta) *metav1.OwnerReference {
	for _, ref := range meta.OwnerReferences {
		if ref.APIVersion == navigator.GroupName+"/v1alpha1" && ref.Kind == kindName {
			return &ref
		}
	}
	return nil
}

func clusterService(c v1alpha1.CouchbaseCluster, name string, http bool, annotations map[string]string, roles ...string) *apiv1.Service {
	svc := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            c.Name + "-" + name,
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Labels:          buildNodePoolLabels(c, "", roles...),
			Annotations:     annotations,
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Name:       "transport",
					Port:       int32(9300),
					TargetPort: intstr.FromInt(9300),
				},
			},
			Selector: buildNodePoolLabels(c, "", roles...),
		},
	}

	if http {
		svc.Spec.Ports = append(svc.Spec.Ports, apiv1.ServicePort{
			Name:       "http",
			Port:       int32(9200),
			TargetPort: intstr.FromInt(9200),
		})
	}

	return &svc
}

func nodePoolIsStateful(np v1alpha1.CouchbaseClusterNodePool) bool {
	return np.Persistence != nil
}

func isManagedByCluster(c v1alpha1.CouchbaseCluster, meta metav1.ObjectMeta) bool {
	clusterOwnerRef := ownerReference(c)
	for _, o := range meta.OwnerReferences {
		if clusterOwnerRef.APIVersion == o.APIVersion &&
			clusterOwnerRef.Kind == o.Kind &&
			clusterOwnerRef.Name == o.Name &&
			clusterOwnerRef.UID == o.UID {
			return true
		}
	}
	return false
}

func nodePoolVersionAnnotation(m map[string]string) string {
	return m[nodePoolVersionAnnotationKey]
}
