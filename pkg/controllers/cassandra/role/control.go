package role

import (
	rbacv1 "k8s.io/api/rbac/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rbaclisters "k8s.io/client-go/listers/rbac/v1beta1"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"

	"k8s.io/client-go/tools/record"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type control struct {
	kubeClient kubernetes.Interface
	roleLister rbaclisters.RoleLister
	recorder   record.EventRecorder
}

var _ Interface = &control{}

func NewControl(
	kubeClient kubernetes.Interface,
	roleLister rbaclisters.RoleLister,
	recorder record.EventRecorder,
) *control {
	return &control{
		kubeClient: kubeClient,
		roleLister: roleLister,
		recorder:   recorder,
	}
}

func (c *control) Sync(cluster *v1alpha1.CassandraCluster) error {
	newRole := RoleForCluster(cluster)
	client := c.kubeClient.RbacV1beta1().Roles(newRole.Namespace)
	existingRole, err := c.roleLister.
		Roles(newRole.Namespace).
		Get(newRole.Name)
	if err == nil {
		return util.OwnerCheck(existingRole, cluster)
	}
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	_, err = client.Create(newRole)
	return err
}

func RoleForCluster(cluster *v1alpha1.CassandraCluster) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.PilotRBACRoleName(cluster),
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				util.NewControllerRef(cluster),
			},
			Labels: util.ClusterLabels(cluster),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs:     []string{"create", "update", "patch"},
				Resources: []string{"events"},
			},
			{
				APIGroups: []string{""},
				Verbs:     []string{"create", "update", "patch", "get", "list", "watch"},
				Resources: []string{"configmaps"},
			},
			{
				APIGroups: []string{""},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"endpoints"},
			},
			{
				APIGroups: []string{navigator.GroupName},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{
					"pilots",
					"cassandraclusters",
				},
			},
			{
				APIGroups: []string{navigator.GroupName},
				Verbs:     []string{"update", "patch"},
				Resources: []string{
					"pilots/status",
				},
			},
		},
	}
}
