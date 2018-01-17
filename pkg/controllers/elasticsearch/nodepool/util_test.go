package nodepool

import (
	"reflect"
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func TestESImageToUse(t *testing.T) {
	type testT struct {
		name          string
		spec          *v1alpha1.ElasticsearchClusterSpec
		expectedImage *v1alpha1.ElasticsearchImage
		expectedError bool
	}
	tests := []testT{
		{
			name: "version specified with no manual image spec",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: "6.1.1",
			},
			expectedImage: &v1alpha1.ElasticsearchImage{
				FsGroup: 1000,
				ImageSpec: v1alpha1.ImageSpec{
					Repository: defaultElasticsearchImageRepository,
					Tag:        "6.1.1",
					PullPolicy: defaultElasticsearchImagePullPolicy,
				},
			},
		},
		{
			name: "version specified with manual image spec",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: "6.1.1",
				Image: &v1alpha1.ElasticsearchImage{
					FsGroup: 1234,
					ImageSpec: v1alpha1.ImageSpec{
						Repository: "abcd",
						Tag:        "xyz",
						PullPolicy: "www",
					},
				},
			},
			expectedImage: &v1alpha1.ElasticsearchImage{
				FsGroup: 1234,
				ImageSpec: v1alpha1.ImageSpec{
					Repository: "abcd",
					Tag:        "xyz",
					PullPolicy: "www",
				},
			},
		},
		{
			name: "no version specified with manual image spec",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: "",
				Image: &v1alpha1.ElasticsearchImage{
					FsGroup: 1234,
					ImageSpec: v1alpha1.ImageSpec{
						Repository: "abcd",
						Tag:        "xyz",
						PullPolicy: "www",
					},
				},
			},
			expectedImage: &v1alpha1.ElasticsearchImage{
				FsGroup: 1234,
				ImageSpec: v1alpha1.ImageSpec{
					Repository: "abcd",
					Tag:        "xyz",
					PullPolicy: "www",
				},
			},
		},
		{
			name: "invalid (non semver) version specified and no manual image",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: "not.semver",
			},
			expectedError: true,
		},
		// We don't error here - if the specified version is non semver but the
		// user has specified a manual image, validation should detect the
		// invalid semver and fail.
		{
			name: "invalid (non semver) version specified with a manual image",
			spec: &v1alpha1.ElasticsearchClusterSpec{
				Version: "not.semver",
				Image: &v1alpha1.ElasticsearchImage{
					FsGroup: 1234,
					ImageSpec: v1alpha1.ImageSpec{
						Repository: "abcd",
						Tag:        "xyz",
						PullPolicy: "www",
					},
				},
			},
			expectedImage: &v1alpha1.ElasticsearchImage{
				FsGroup: 1234,
				ImageSpec: v1alpha1.ImageSpec{
					Repository: "abcd",
					Tag:        "xyz",
					PullPolicy: "www",
				},
			},
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
