package role

import (
	"testing"
	"time"

	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	navigator "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	kubeClient *kubefake.Clientset
	// Objects to put in the store.
	roleLister []*rbac.Role

	// Actions expected to happen on the client.
	kubeActions []core.Action

	// Objects from here preloaded into NewSimpleFake.
	kubeObjects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.kubeObjects = []runtime.Object{}
	return f
}

func newESCluster(name string) *navigator.ElasticsearchCluster {
	return &navigator.ElasticsearchCluster{
		TypeMeta: metav1.TypeMeta{APIVersion: navigator.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			UID:       uuid.NewUUID(),
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: navigator.ElasticsearchClusterSpec{},
	}
}

func (f *fixture) newController() (Interface, kubeinformers.SharedInformerFactory) {
	f.kubeClient = kubefake.NewSimpleClientset(f.kubeObjects...)

	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeClient, noResyncPeriodFunc())

	c := NewController(f.kubeClient, k8sI.Rbac().V1beta1().Roles().Lister(), nil)

	for _, f := range f.roleLister {
		k8sI.Rbac().V1beta1().Roles().Informer().GetIndexer().Add(f)
	}

	return c, k8sI
}

func (f *fixture) run(cluster *navigator.ElasticsearchCluster) {
	f.runController(cluster, true, false)
}

func (f *fixture) runExpectError(cluster *navigator.ElasticsearchCluster) {
	f.runController(cluster, true, true)
}

func (f *fixture) runController(cluster *navigator.ElasticsearchCluster, startInformers bool, expectError bool) {
	c, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.Sync(cluster)
	if !expectError && err != nil {
		f.t.Errorf("error syncing ElasticsearchCluster: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing ElasticsearchCluster, got nil")
	}

	k8sActions := filterInformerActions(f.kubeClient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeActions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeActions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeActions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) && action.GetSubresource() == expectedAction.GetSubresource()) {
			f.t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.kubeActions) > len(k8sActions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeActions)-len(k8sActions), f.kubeActions[len(k8sActions):])
	}
}

func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "foos") ||
				action.Matches("watch", "foos") ||
				action.Matches("list", "deployments") ||
				action.Matches("watch", "deployments")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectCreateRoleAction(d *rbac.Role) {
	f.kubeActions = append(f.kubeActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "roles"}, d.Namespace, d))
}

func (f *fixture) expectUpdateRoleAction(d *rbac.Role) {
	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "roles"}, d.Namespace, d))
}

func TestCreatesRole(t *testing.T) {
	f := newFixture(t)
	cluster := newESCluster("test")
	role := roleForCluster(cluster)

	f.expectCreateRoleAction(role)

	// TODO: invent some way to neatly check the status of the 'cluster' after running
	f.run(cluster)
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	cluster := newESCluster("test")
	role := roleForCluster(cluster)

	f.roleLister = append(f.roleLister, role)
	f.kubeObjects = append(f.kubeObjects, role)

	// TODO: invent some way to neatly check the status of the 'cluster' after running
	f.run(cluster)
}

func TestUpdateRole(t *testing.T) {
	f := newFixture(t)
	cluster := newESCluster("test")
	role := roleForCluster(cluster)
	role.Rules = []rbac.PolicyRule{}
	f.roleLister = append(f.roleLister, role)
	f.kubeObjects = append(f.kubeObjects, role)

	// TODO: make this actually verify that the role now has the correct PolicyRules
	f.expectUpdateRoleAction(role)
	// TODO: invent some way to neatly check the status of the 'cluster' after running
	f.run(cluster)
}

func TestNotControlledByUs(t *testing.T) {
	f := newFixture(t)
	cluster := newESCluster("test")
	role := roleForCluster(cluster)
	role.ObjectMeta.OwnerReferences = []metav1.OwnerReference{}

	f.roleLister = append(f.roleLister, role)
	f.kubeObjects = append(f.kubeObjects, role)

	// TODO: invent some way to neatly check the status of the 'cluster' after running
	f.runExpectError(cluster)
}

func int32Ptr(i int32) *int32 { return &i }
