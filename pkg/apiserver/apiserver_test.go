package apiserver_test

import (
	"net"
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/apiserver"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/server"
	storagetesting "k8s.io/apiserver/pkg/storage/etcd/testing"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
)

var (
	v1GroupVersion = schema.GroupVersion{Group: "", Version: "v1"}

	scheme         = runtime.NewScheme()
	codecs         = serializer.NewCodecFactory(scheme)
	parameterCodec = runtime.NewParameterCodec(scheme)
)

func TestApiServer(t *testing.T) {
	scheme := runtime.NewScheme()
	etcdServer, _ := storagetesting.NewUnsecuredEtcd3TestClientServer(t, scheme)
	defer etcdServer.Terminate(t)
	config := server.NewConfig(codecs)
	config.PublicAddress = net.ParseIP("192.168.10.4")
	config.RequestContextMapper = genericapirequest.NewRequestContextMapper()
	config.LegacyAPIGroupPrefixes = sets.NewString("/api")
	config.LoopbackClientConfig = &restclient.Config{}

	clientset := fake.NewSimpleClientset()
	if clientset == nil {
		t.Fatal("unable to create fake client set")
	}
	config.SharedInformerFactory = informers.NewSharedInformerFactory(clientset, config.LoopbackClientConfig.Timeout)
	config.SwaggerConfig = server.DefaultSwaggerConfig()

	apiConfig := &apiserver.Config{
		GenericConfig: config,
	}

	s, err := apiConfig.Complete().New()
	if err != nil {
		t.Fatal(err)
	}

	stopCh := make(chan struct{})
	err = s.GenericAPIServer.PrepareRun().Run(stopCh)
	if err != nil {
		t.Fatal(err)
	}
}
