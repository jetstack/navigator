package testing

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1beta1"

	navinformers "github.com/jetstack/navigator/pkg/client/informers/externalversions"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/role"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/rolebinding"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/seedlabeller"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/serviceaccount"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	navigatorfake "github.com/jetstack/navigator/pkg/client/clientset/versioned/fake"
)

func ClusterForTest() *v1alpha1.CassandraCluster {
	c := &v1alpha1.CassandraCluster{
		Spec: v1alpha1.CassandraClusterSpec{
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
	t                          *testing.T
	Cluster                    *v1alpha1.CassandraCluster
	SeedProviderServiceControl cassandra.ControlInterface
	NodesServiceControl        cassandra.ControlInterface
	NodepoolControl            nodepool.Interface
	PilotControl               pilot.Interface
	ServiceAccountControl      serviceaccount.Interface
	RoleControl                role.Interface
	RoleBindingControl         rolebinding.Interface
	SeedLabellerControl        seedlabeller.Interface
	k8sClient                  *fake.Clientset
	k8sObjects                 []runtime.Object
	naviClient                 *navigatorfake.Clientset
	naviObjects                []runtime.Object
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

func (f *Fixture) AddObjectN(o runtime.Object) {
	f.naviObjects = append(f.naviObjects, o)
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
	if f.SeedProviderServiceControl == nil {
		f.SeedProviderServiceControl = service.NewControl(
			f.k8sClient,
			services,
			recorder,
			service.SeedsServiceForCluster,
		)
	}
	if f.NodesServiceControl == nil {
		f.NodesServiceControl = service.NewControl(
			f.k8sClient,
			services,
			recorder,
			service.NodesServiceForCluster,
		)
	}
	statefulSets := k8sFactory.Apps().V1beta1().StatefulSets().Lister()
	pods := k8sFactory.Core().V1().Pods().Lister()
	if f.NodepoolControl == nil {
		f.NodepoolControl = nodepool.NewControl(
			f.k8sClient,
			statefulSets,
			recorder,
		)
	}
	f.naviClient = navigatorfake.NewSimpleClientset(f.naviObjects...)
	naviFactory := navinformers.NewSharedInformerFactory(f.naviClient, 0)
	pilots := naviFactory.Navigator().V1alpha1().Pilots().Lister()
	if f.PilotControl == nil {
		f.PilotControl = pilot.NewControl(
			f.naviClient,
			pilots,
			pods,
			statefulSets,
			recorder,
		)
	}
	serviceAccounts := k8sFactory.Core().V1().ServiceAccounts().Lister()
	if f.ServiceAccountControl == nil {
		f.ServiceAccountControl = serviceaccount.NewControl(
			f.k8sClient,
			serviceAccounts,
			recorder,
		)
	}

	roles := k8sFactory.Rbac().V1beta1().Roles().Lister()
	if f.RoleControl == nil {
		f.RoleControl = role.NewControl(
			f.k8sClient,
			roles,
			recorder,
		)
	}

	roleBindings := k8sFactory.Rbac().V1beta1().RoleBindings().Lister()
	if f.RoleBindingControl == nil {
		f.RoleBindingControl = rolebinding.NewControl(
			f.k8sClient,
			roleBindings,
			recorder,
		)
	}

	if f.SeedLabellerControl == nil {
		f.SeedLabellerControl = seedlabeller.NewControl(
			f.k8sClient,
			statefulSets,
			pods,
			recorder,
		)
	}

	c := cassandra.NewControl(
		f.SeedProviderServiceControl,
		f.NodesServiceControl,
		f.NodepoolControl,
		f.PilotControl,
		f.ServiceAccountControl,
		f.RoleControl,
		f.RoleBindingControl,
		f.SeedLabellerControl,
		recorder,
	)
	stopCh := make(chan struct{})
	defer close(stopCh)
	k8sFactory.Start(stopCh)
	naviFactory.Start(stopCh)
	if !cache.WaitForCacheSync(
		stopCh,
		k8sFactory.Core().V1().Pods().Informer().HasSynced,
		k8sFactory.Core().V1().Services().Informer().HasSynced,
		k8sFactory.Apps().V1beta1().StatefulSets().Informer().HasSynced,
		naviFactory.Navigator().V1alpha1().Pilots().Informer().HasSynced,
		k8sFactory.Core().V1().ServiceAccounts().Informer().HasSynced,
		k8sFactory.Rbac().V1beta1().Roles().Informer().HasSynced,
		k8sFactory.Rbac().V1beta1().RoleBindings().Informer().HasSynced,
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

func (f *Fixture) ServiceAccounts() *v1.ServiceAccountList {
	serviceAccounts, err := f.k8sClient.
		CoreV1().
		ServiceAccounts(f.Cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return serviceAccounts
}

func (f *Fixture) AssertServiceAccountsLength(l int) {
	serviceAccounts := f.ServiceAccounts()
	serviceAccountsLength := len(serviceAccounts.Items)
	if serviceAccountsLength != l {
		f.t.Log(serviceAccounts)
		f.t.Errorf(
			"Incorrect number of services accounts. Expected %d. Got %d.",
			l,
			serviceAccountsLength,
		)
	}
}

func (f *Fixture) Roles() *rbacv1.RoleList {
	roles, err := f.k8sClient.
		RbacV1beta1().
		Roles(f.Cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return roles
}

func (f *Fixture) AssertRolesLength(l int) {
	roles := f.Roles()
	rolesLength := len(roles.Items)
	if rolesLength != l {
		f.t.Log(roles)
		f.t.Errorf(
			"Incorrect number of roles. Expected %d. Got %d.",
			l,
			rolesLength,
		)
	}
}

func (f *Fixture) RoleBindings() *rbacv1.RoleBindingList {
	roleBindings, err := f.k8sClient.
		RbacV1beta1().
		RoleBindings(f.Cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return roleBindings
}

func (f *Fixture) AssertRoleBindingsLength(l int) {
	roleBindings := f.RoleBindings()
	roleBindingsLength := len(roleBindings.Items)
	if roleBindingsLength != l {
		f.t.Log(roleBindings)
		f.t.Errorf(
			"Incorrect number of role bindings. Expected %d. Got %d.",
			l,
			roleBindingsLength,
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

func (f *Fixture) Pilots() *v1alpha1.PilotList {
	pilots, err := f.naviClient.
		NavigatorV1alpha1().
		Pilots(f.Cluster.Namespace).
		List(metav1.ListOptions{})
	if err != nil {
		f.t.Fatal(err)
	}
	return pilots
}

func (f *Fixture) AssertPilotsLength(l int) {
	sets := f.Pilots()
	setsLength := len(sets.Items)
	if setsLength != l {
		f.t.Log(sets)
		f.t.Errorf(
			"Incorrect number of Pilots: %#v", setsLength,
		)
	}
}

type FakeControl struct {
	SyncError error
}

func (c *FakeControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	return c.SyncError
}
