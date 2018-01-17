package role

import (
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

func roleForCluster(c *v1alpha1.ElasticsearchCluster) *rbacv1beta1.Role {
	return &rbacv1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.RBACRoleName(c),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
			Labels:          util.ClusterLabels(c),
		},
		Rules: []rbacv1beta1.PolicyRule{
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
				APIGroups: []string{navigator.GroupName},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{
					"pilots",
					"elasticsearchclusters",
				},
			},
			{
				APIGroups: []string{navigator.GroupName},
				Verbs:     []string{"update", "patch"},
				Resources: []string{
					"pilots/status",
					"elasticsearchclusters/status",
				},
			},
		},
	}
}
