package tpr

import (
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/jetstack-experimental/navigator/pkg/api/v1"
)

var thirdPartyResource = &v1beta1.ThirdPartyResource{
	ObjectMeta: metav1.ObjectMeta{
		Name: "elasticsearch-cluster." + v1.GroupName,
	},
	Description: "A specification of an Elasticsearch cluster",
	Versions: []v1beta1.APIVersion{
		{
			Name: "v1",
		},
	},
}

// Ensure will ensure that the appropriate ThirdPartyResources exist in
// the target Kubernetes cluster
func Ensure(cl *kubernetes.Clientset) error {
	_, err := cl.Extensions().ThirdPartyResources().Create(thirdPartyResource)

	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
