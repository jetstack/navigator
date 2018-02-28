package nodepool

import (
	"github.com/coreos/go-semver/semver"
	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func cassImageToUse(spec *v1alpha1.CassandraClusterSpec) *v1alpha1.ImageSpec {
	if spec.Image == nil {
		return defaultCassandraImageForVersion(spec.Version)
	}
	return spec.Image

}

const defaultCassandraImageRepository = "docker.io/cassandra"
const defaultCassandraImagePullPolicy = corev1.PullIfNotPresent

func defaultCassandraImageForVersion(v semver.Version) *v1alpha1.ImageSpec {
	return &v1alpha1.ImageSpec{
		Repository: defaultCassandraImageRepository,
		Tag:        v.String(),
		PullPolicy: defaultCassandraImagePullPolicy,
	}
}
