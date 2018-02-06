package actions

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func esImageToUse(spec *v1alpha1.ElasticsearchClusterSpec) (*v1alpha1.ImageSpec, error) {
	if spec.Image == nil {
		return defaultElasticsearchImageForVersion(spec.Version)
	}
	return spec.Image, nil
}

const defaultElasticsearchImageRepository = "docker.elastic.co/elasticsearch/elasticsearch"
const defaultElasticsearchRunAsUser = 1000
const defaultElasticsearchImagePullPolicy = corev1.PullIfNotPresent

func defaultElasticsearchImageForVersion(v semver.Version) (*v1alpha1.ImageSpec, error) {
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		return nil, fmt.Errorf("version must be specified")
	}
	return &v1alpha1.ImageSpec{
		Repository: defaultElasticsearchImageRepository,
		Tag:        v.String(),
		PullPolicy: defaultElasticsearchImagePullPolicy,
	}, nil
}
