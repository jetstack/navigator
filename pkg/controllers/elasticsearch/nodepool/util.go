package nodepool

import (
	"github.com/coreos/go-semver/semver"
	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func esImageToUse(spec *v1alpha1.ElasticsearchClusterSpec) (*v1alpha1.ElasticsearchImage, error) {
	if spec.Image == nil {
		return defaultElasticsearchImageForVersion(spec.Version)
	}
	return spec.Image, nil
}

const defaultElasticsearchImageRepository = "docker.elastic.co/elasticsearch/elasticsearch"
const defaultElasticsearchRunAsUser = 1000
const defaultElasticsearchImagePullPolicy = string(corev1.PullIfNotPresent)

func defaultElasticsearchImageForVersion(s string) (*v1alpha1.ElasticsearchImage, error) {
	// ensure the version follows semver
	_, err := semver.NewVersion(s)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.ElasticsearchImage{
		FsGroup: defaultElasticsearchRunAsUser,
		ImageSpec: v1alpha1.ImageSpec{
			Repository: defaultElasticsearchImageRepository,
			Tag:        s,
			PullPolicy: defaultElasticsearchImagePullPolicy,
		},
	}, nil
}
