package controllers

import (
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
)

func PodControlledByCluster(
	cluster metav1.Object,
	pod *apiv1.Pod,
	ssLister appslisters.StatefulSetLister,
) (bool, error) {
	if metav1.IsControlledBy(pod, cluster) {
		return true, nil
	}
	ownerRef := metav1.GetControllerOf(pod)
	if ownerRef == nil {
		return false, nil
	}
	if ownerRef.Kind != "StatefulSet" {
		return false, nil
	}
	set, err := ssLister.StatefulSets(pod.Namespace).Get(ownerRef.Name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return metav1.IsControlledBy(set, cluster), nil
}
