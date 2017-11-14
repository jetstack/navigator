package testing

import (
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

func ClusterForTest() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{
		Spec: v1alpha1.CassandraClusterSpec{
			CqlPort: 9042,
			NodePools: []v1alpha1.CassandraClusterNodePool{
				v1alpha1.CassandraClusterNodePool{
					Name:     "RingNodes",
					Replicas: 3,
				},
			},
		},
	}
	c.SetName("cassandra-1")
	c.SetNamespace("app-1")
	return c
}

type Fixture struct {
	t               *testing.T
	Cluster         *v1alpha1.CassandraCluster
	ServiceControl  service.Interface
	NodepoolControl nodepool.Interface
	k8sClient       *fake.Clientset
	k8sObjects      []runtime.Object
}

func NewFixture(t *testing.T) *Fixture {
	return &Fixture{
		t:       t,
		Cluster: ClusterForTest(),
	}
}

func (f *Fixture) AddObjectK(o runtime.Object) {
	f.k8sObjects = append(f.k8sObjects, o)
}

func (f *Fixture) setupAndSync() error {
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
	if f.ServiceControl == nil {
		f.ServiceControl = service.NewControl(f.k8sClient, services, recorder)
	}

	statefulSets := k8sFactory.Apps().V1beta1().StatefulSets().Lister()
	if f.NodepoolControl == nil {
		f.NodepoolControl = nodepool.NewControl(
			f.k8sClient,
			statefulSets,
			recorder,
		)
	}

	c := cassandra.NewControl(
		f.ServiceControl,
		f.NodepoolControl,
		recorder,
	)
	stopCh := make(chan struct{})
	defer close(stopCh)
	k8sFactory.Start(stopCh)
	if !cache.WaitForCacheSync(
		stopCh,
		k8sFactory.Core().V1().Services().Informer().HasSynced,
		k8sFactory.Apps().V1beta1().StatefulSets().Informer().HasSynced,
	) {
		f.t.Fatal("WaitForCacheSync failure")
	}
	return c.Sync(f.Cluster)
}

func (f *Fixture) Run() {
	err := f.setupAndSync()
	if err != nil {
		f.t.Error(err)
	}
}

func (f *Fixture) RunExpectError() {
	err := f.setupAndSync()
	if err == nil {
		f.t.Error("Sync was expected to return an error. Got nil.")
	}
}

func (f *Fixture) Services() *v1.ServiceList {
	services, err := f.k8sClient.
		CoreV1().
		Services(f.Cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return services
}

func (f *Fixture) AssertServicesLength(l int) {
	services := f.Services()
	servicesLength := len(services.Items)
	if servicesLength != l {
		f.t.Log(services)
		f.t.Errorf(
			"Incorrect number of services: %#v", servicesLength,
		)
	}
}

func (f *Fixture) StatefulSets() *apps.StatefulSetList {
	sets, err := f.k8sClient.
		AppsV1beta1().
		StatefulSets(f.Cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return sets
}

func (f *Fixture) AssertStatefulSetsLength(l int) {
	sets := f.StatefulSets()
	setsLength := len(sets.Items)
	if setsLength != l {
		f.t.Log(sets)
		f.t.Errorf(
			"Incorrect number of StatefulSets: %#v", setsLength,
		)
	}
}

type FakeControl struct {
	SyncError error
}

func (c *FakeControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	return c.SyncError
}
