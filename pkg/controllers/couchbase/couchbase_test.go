package couchbase_test

import (
	"testing"

	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/fake"
	intinformers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	"github.com/jetstack-experimental/navigator/pkg/controllers/couchbase"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCouchbaseController(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	navigatorClient := navigatorfake.NewSimpleClientset()
	ctx := controllers.Context{
		Client:                   kubeClient,
		InformerFactory:          informers.NewSharedInformerFactory(kubeClient, 0),
		NavigatorInformerFactory: intinformers.NewSharedInformerFactory(navigatorClient, 0),
		Namespace:                metav1.NamespaceAll,
		Stop:                     make(<-chan struct{}),
	}

	controller := couchbase.NewCouchbase(
		ctx.NavigatorInformerFactory.Navigator().V1alpha1().CouchbaseClusters(),
		nil,
	)
	controller.Run(1, ctx.Stop)
}
