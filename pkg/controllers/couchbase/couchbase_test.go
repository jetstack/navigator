package couchbase_test

import (
	"testing"
	"time"

	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	intinformers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCouchbaseController(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	navigatorClient := navigatorfake.NewSimpleClientset()
	stop := make(chan struct{})
	ctx := &controllers.Context{
		Client:                   kubeClient,
		InformerFactory:          informers.NewSharedInformerFactory(kubeClient, 0),
		NavigatorInformerFactory: intinformers.NewSharedInformerFactory(navigatorClient, 0),
		Namespace:                metav1.NamespaceAll,
		Stop:                     stop,
	}

	go func() {
		<-time.After(time.Second * 5)
		close(stop)
	}()

	err := controllers.Start(ctx, controllers.Known(), stop)
	if err != nil {
		t.Fatal(err)
	}
}
