package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// TODO: update to quay.io
	defaultPilotImageRepository = "jetstackexperimental/navigator-pilot-elasticsearch"
	// TODO: don't use latest
	defaultPilotImageTag = "latest"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_ElasticsearchPilotImage(obj *ElasticsearchPilotImage) {
	if obj.PullPolicy == "" {
		obj.PullPolicy = corev1.PullIfNotPresent
	}
	if obj.Repository == "" {
		obj.Repository = defaultPilotImageRepository
	}
	if obj.Tag == "" {
		obj.Tag = defaultPilotImageTag
	}
}
