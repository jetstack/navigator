package cassandra

import (
	"testing"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

type Fixture struct {
	t          *testing.T
	cluster    *v1alpha1.CassandraCluster
	k8sClient  *fake.Clientset
	k8sObjects []runtime.Object
}

func NewFixture(t *testing.T) *Fixture {
	return &Fixture{
		t: t,
	}
}

func (f *Fixture) Run(cluster *v1alpha1.CassandraCluster) {
	f.cluster = cluster
	recorder := record.NewFakeRecorder(0)
	finished := make(chan struct{})
	defer func() {
		close(recorder.Events)
		<-finished
	}()
	go func() {
		for e := range recorder.Events {
			f.t.Logf("EVENT: %q", e)
		}
		close(finished)
	}()
	f.k8sClient = fake.NewSimpleClientset(f.k8sObjects...)
	k8sFactory := informers.NewSharedInformerFactory(f.k8sClient, 0)
	services := k8sFactory.Core().V1().Services().Lister()
	c := NewControl(
		service.NewControl(f.k8sClient, services, recorder),
		recorder,
	)
	err := c.Sync(cluster)
	if err != nil {
		f.t.Error(err)
	}
}

func (f *Fixture) services() *v1.ServiceList {
	services, err := f.k8sClient.
		CoreV1().
		Services(f.cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return services
}

func (f *Fixture) assertServicesLength(l int) {
	services := f.services()
	servicesLength := len(services.Items)
	if servicesLength != 1 {
		f.t.Log(services)
		f.t.Errorf(
			"Incorrect number of services: %#v", servicesLength,
		)
	}
}

func TestServiceSync(t *testing.T) {
	t.Run(
		"service created",
		func(t *testing.T) {
			f := NewFixture(t)
			cluster := newCassandraCluster()
			f.Run(cluster)
			f.assertServicesLength(1)
		},
	)
	t.Run(
		"service exists",
		func(t *testing.T) {
			f := NewFixture(t)
			cluster := newCassandraCluster()
			f.k8sObjects = append(
				f.k8sObjects,
				service.ServiceForCluster(cluster),
			)
			f.Run(cluster)
			f.assertServicesLength(1)
		},
	)
	t.Run(
		"service needs sync",
		func(t *testing.T) {
			f := NewFixture(t)
			cluster := newCassandraCluster()
			// Remove the ports from the default cluster and expect them to be
			// re-created.
			unsyncedService := service.ServiceForCluster(cluster)
			unsyncedService.Spec.Ports = []v1.ServicePort{}
			f.k8sObjects = append(
				f.k8sObjects,
				unsyncedService,
			)
			f.Run(cluster)
			services := f.services()
			service := services.Items[0]
			if len(service.Spec.Ports) == 0 {
				t.Log(service)
				f.t.Error("Service was not updated")
			}
		},
	)
}
