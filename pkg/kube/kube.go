package kube

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/jetstack-experimental/navigator/pkg/apis/marshal"
)

var thirdPartyResource = &v1beta1.ThirdPartyResource{
	ObjectMeta: metav1.ObjectMeta{
		Name: "elasticsearch-cluster." + marshal.GroupName,
	},
	Description: "A specification of an Elasticsearch cluster",
	Versions: []v1beta1.APIVersion{
		{
			Name: "v1alpha1",
		},
	},
}

// Config will return a rest.Config for communicating with the Kubernetes API server.
// If apiServerHost is specified, a config without authentication that is configured
// to talk to the apiServerHost URL will be returned. Else, the in-cluster config will be loaded,
// and failing this, the config will be loaded from the users local kubeconfig directory
func Config(apiServerHost string) (*rest.Config, error) {
	var err error
	var cfg *rest.Config

	if len(apiServerHost) > 0 {
		cfg = new(rest.Config)
		cfg.Host = apiServerHost
	} else if cfg, err = rest.InClusterConfig(); err != nil {
		apiCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()

		if err != nil {
			return nil, fmt.Errorf("error loading cluster config: %s", err.Error())
		}

		cfg, err = clientcmd.NewDefaultClientConfig(*apiCfg, &clientcmd.ConfigOverrides{}).ClientConfig()

		if err != nil {
			return nil, fmt.Errorf("error loading cluster client config: %s", err.Error())
		}
	}

	return cfg, nil
}

// EnsureTPR will ensure that the appropriate ThirdPartyResources exist in
// the target Kubernetes cluster
func EnsureTPR(cl *kubernetes.Clientset) error {
	_, err := cl.Extensions().ThirdPartyResources().Create(thirdPartyResource)

	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
