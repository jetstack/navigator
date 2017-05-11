package kube

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/jetstack-experimental/navigator/pkg/apis/marshal"
)

func NewMarshalRESTClient(apiServerHost string) (*rest.RESTClient, error) {
	cfg, err := Config(apiServerHost)

	if err != nil {
		return nil, fmt.Errorf("error loading kubernetes client config: %s", err.Error())
	}

	configureTprClient(cfg)

	return rest.RESTClientFor(cfg)
}

// NewKubernetesClient will return an authenticated Kubernetes client.
// If apiServerHost is specified, a config without authentication that is configured
// to talk to the apiServerHost URL will be returned. Else, the in-cluster config will be loaded,
// and failing this, the config will be loaded from the users local kubeconfig directory
func NewKubernetesClient(apiServerHost string) (*kubernetes.Clientset, error) {
	cfg, err := Config(apiServerHost)

	if err != nil {
		return nil, fmt.Errorf("error loading kubernetes client config: %s", err.Error())
	}

	cl, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, fmt.Errorf("error instantiating kubernetes client connection: %s", err.Error())
	}

	return cl, nil
}

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

func configureTprClient(config *rest.Config) {
	config.GroupVersion = &marshal.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
}
