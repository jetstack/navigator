package rolebinding

import (
	rbacv1 "k8s.io/api/rbac/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rbaclisters "k8s.io/client-go/listers/rbac/v1beta1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"

	"k8s.io/client-go/tools/record"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type control struct {
	kubeClient        kubernetes.Interface
	roleBindingLister rbaclisters.RoleBindingLister
	recorder          record.EventRecorder
}

var _ Interface = &control{}

func NewControl(
	kubeClient kubernetes.Interface,
	roleBindingLister rbaclisters.RoleBindingLister,
	recorder record.EventRecorder,
) *control {
	return &control{
		kubeClient:        kubeClient,
		roleBindingLister: roleBindingLister,
		recorder:          recorder,
	}
}

func (c *control) Sync(cluster *v1alpha1.CassandraCluster) error {
	newRoleBinding := RoleBindingForCluster(cluster)
	client := c.kubeClient.RbacV1beta1().RoleBindings(newRoleBinding.Namespace)
	existingRoleBinding, err := c.roleBindingLister.
		RoleBindings(newRoleBinding.Namespace).
		Get(newRoleBinding.Name)
	if err == nil {
		return util.OwnerCheck(existingRoleBinding, cluster)
	}
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	_, err = client.Create(newRoleBinding)
	return err
}

func RoleBindingForCluster(cluster *v1alpha1.CassandraCluster) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.PilotRBACRoleName(cluster),
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				util.NewControllerRef(cluster),
			},
			Labels: util.ClusterLabels(cluster),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: rbacv1.ServiceAccountKind,
				Name: util.ServiceAccountName(cluster),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: util.PilotRBACRoleName(cluster),
		},
	}
}
