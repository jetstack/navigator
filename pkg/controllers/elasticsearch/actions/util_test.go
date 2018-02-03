package actions

import (
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

var validESVersion = semver.New("6.1.1")

func TestESImageToUse(t *testing.T) {
	type testT struct {
		name          string
		spec          *v1alpha1.ElasticsearchClusterSpec
		expectedImage *v1alpha1.ImageSpec
		expectedError bool
	}
	tests := []testT{
		{
			name: "version specified with no manual image spec",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: *validESVersion,
			},
			expectedImage: &v1alpha1.ImageSpec{
				Repository: defaultElasticsearchImageRepository,
				Tag:        "6.1.1",
				PullPolicy: defaultElasticsearchImagePullPolicy,
			},
		},
		{
			name: "version specified with manual image spec",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: *validESVersion,
				Image: &v1alpha1.ImageSpec{
					Repository: "abcd",
					Tag:        "xyz",
					PullPolicy: defaultElasticsearchImagePullPolicy,
				},
			},
			expectedImage: &v1alpha1.ImageSpec{
				Repository: "abcd",
				Tag:        "xyz",
				PullPolicy: defaultElasticsearchImagePullPolicy,
			},
		},
		{
			name: "no version specified with manual image spec",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Image: &v1alpha1.ImageSpec{
					Repository: "abcd",
					Tag:        "xyz",
					PullPolicy: defaultElasticsearchImagePullPolicy,
				},
			},
			expectedImage: &v1alpha1.ImageSpec{
				Repository: "abcd",
				Tag:        "xyz",
				PullPolicy: defaultElasticsearchImagePullPolicy,
			},
		},
		{
			name:          "no version specified and no manual image",
			spec:          &v1alpha1.ElasticsearchClusterSpec{},
			expectedError: true,
		},
	}
	testF := func(test testT) func(*testing.T) {
		return func(t *testing.T) {
			img, err := esImageToUse(test.spec)
			if err != nil && !test.expectedError {
				t.Errorf("expected no error but got: %s", err)
			}
			if err == nil && test.expectedError {
				t.Errorf("expected error to be returned but got none")
			}
			if !reflect.DeepEqual(img, test.expectedImage) {
				t.Errorf("expected %+v to equal %+v", img, test.expectedImage)
			}
		}
	}
	for _, test := range tests {
		t.Run(test.name, testF(test))
	}
}
