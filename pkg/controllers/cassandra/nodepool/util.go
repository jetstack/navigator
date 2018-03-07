package nodepool

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/cassandra/version"
)

func CassImageToUse(spec *v1alpha1.CassandraClusterSpec) *v1alpha1.ImageSpec {
	if spec.Image == nil {
		return defaultCassandraImageForVersion(spec.Version)
	}
	return spec.Image

}

const defaultCassandraImageRepository = "docker.io/cassandra"
const defaultCassandraImagePullPolicy = corev1.PullIfNotPresent

func defaultCassandraImageForVersion(v version.Version) *v1alpha1.ImageSpec {
	return &v1alpha1.ImageSpec{
		Repository: defaultCassandraImageRepository,
		Tag:        v.Semver(),
		PullPolicy: defaultCassandraImagePullPolicy,
	}
}
