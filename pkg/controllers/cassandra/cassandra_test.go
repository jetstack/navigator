package cassandra

import (
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	navigatorfake "github.com/jetstack-experimental/navigator/pkg/client/clientset/versioned/fake"
	"github.com/jetstack-experimental/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type fixture struct {
	t *testing.T

	client    *navigatorfake.Clientset
	k8sClient *fake.Clientset

	cassandraClusters []*v1alpha1.CassandraCluster
	services          []*corev1.Service

	objects    []runtime.Object
	k8sObjects []runtime.Object

	actions    []clienttesting.Action
	k8sActions []clienttesting.Action
}

func newFixture(t *testing.T) *fixture {
	return &fixture{
		t: t,
	}
}

func (f *fixture) newController() *CassandraController {
	nclient := navigatorfake.NewSimpleClientset(f.objects...)
	f.client = nclient
	nfactory := externalversions.NewSharedInformerFactory(nclient, 0)
	cassClusters := nfactory.Navigator().V1alpha1().CassandraClusters().Informer()
	for _, c := range f.cassandraClusters {
		err := cassClusters.GetIndexer().Add(c)
		if err != nil {
			f.t.Error(err)
		}
	}

	kclient := fake.NewSimpleClientset(f.k8sObjects...)
	f.k8sClient = kclient
	kfactory := informers.NewSharedInformerFactory(kclient, 0)
	services := kfactory.Core().V1().Services().Informer()
	for _, s := range f.services {
		err := services.GetIndexer().Add(s)
		if err != nil {
			f.t.Error(err)
		}
	}

	recorder := &record.FakeRecorder{}

	return NewCassandra(
		nclient,
		kclient,
		cassClusters,
		services,
		recorder,
	)
}

func (f *fixture) run(key string) {
	c := f.newController()
	err := c.sync(key)
	if err != nil {
		f.t.Error(err)
	}
	f.verifyActions()
}

func (f *fixture) verifyActions() {
	actions := f.client.Actions()
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf(
				"%d unexpected actions: %+v", len(actions)-len(f.actions),
				actions[i:],
			)
			break
		}

		expectedAction := f.actions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) &&
			action.GetSubresource() == expectedAction.GetSubresource()) {
			f.t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf(
			"%d additional expected actions:%+v",
			len(f.actions)-len(actions), f.actions[len(actions):],
		)
	}

	k8sActions := f.k8sClient.Actions()
	for i, action := range k8sActions {
		if len(f.k8sActions) < i+1 {
			f.t.Errorf(
				"%d unexpected actions: %+v",
				len(k8sActions)-len(f.k8sActions), k8sActions[i:],
			)
			break
		}

		expectedAction := f.k8sActions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) &&
			action.GetSubresource() == expectedAction.GetSubresource()) {
			f.t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.k8sActions) > len(k8sActions) {
		f.t.Errorf(
			"%d additional expected actions:%+v",
			len(f.k8sActions)-len(k8sActions), f.k8sActions[len(k8sActions):],
		)
	}
}

func (f *fixture) addCassandraCluster(cluster *v1alpha1.CassandraCluster) {
	f.objects = append(f.objects, cluster)
	f.cassandraClusters = append(f.cassandraClusters, cluster)
}

func (f *fixture) addService(service *corev1.Service) {
	f.k8sObjects = append(f.k8sObjects, service)
	f.services = append(f.services, service)
}

func (f *fixture) expectCreateServiceAction(s *corev1.Service) {
	f.k8sActions = append(
		f.k8sActions,
		clienttesting.NewCreateAction(
			schema.GroupVersionResource{Resource: "services"},
			s.Namespace, s,
		),
	)
}

func (f *fixture) expectUpdateServiceAction(s *corev1.Service) {
	f.k8sActions = append(
		f.k8sActions,
		clienttesting.NewUpdateAction(
			schema.GroupVersionResource{Resource: "services"},
			s.Namespace, s,
		),
	)
}

func newCassandraCluster() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{}
	c.SetName("cassandra-1")
	c.SetNamespace("app-1")
	return c
}

func getKey(t *testing.T, cluster *v1alpha1.CassandraCluster) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cluster)
	if err != nil {
		t.Errorf("Unexpected error getting key for cluster %v: %v", cluster.Name, err)
		return ""
	}
	return key
}

func TestCassandraControllerSync(t *testing.T) {
	t.Run(
		"Services are created",
		func(t *testing.T) {
			cluster := newCassandraCluster()
			f := newFixture(t)
			f.addCassandraCluster(cluster)
			expectedService := service.ServiceForCluster(cluster)
			f.expectCreateServiceAction(expectedService)
			f.run(getKey(t, cluster))
		},
	)
	t.Run(
		"Services are updated",
		func(t *testing.T) {
			cluster := newCassandraCluster()
			f := newFixture(t)
			f.addCassandraCluster(cluster)
			originalService := service.ServiceForCluster(cluster)
			f.addService(originalService)
			expectedService := service.ServiceForCluster(cluster)
			f.expectUpdateServiceAction(expectedService)
			f.run(getKey(t, cluster))
		},
	)
}
