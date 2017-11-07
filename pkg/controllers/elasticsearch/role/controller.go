package role

import (
	"fmt"
	"reflect"

	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rbaclisters "k8s.io/client-go/listers/rbac/v1beta1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

type Interface interface {
	Sync(*v1alpha1.ElasticsearchCluster) error
}

type roleControl struct {
	kubeClient kubernetes.Interface
	roleLister rbaclisters.RoleLister

	recorder record.EventRecorder
}

var _ Interface = &roleControl{}

func NewController(
	kubeClient kubernetes.Interface,
	roleLister rbaclisters.RoleLister,
	recorder record.EventRecorder,
) Interface {
	return &roleControl{
		kubeClient: kubeClient,
		roleLister: roleLister,
		recorder:   recorder,
	}
}

func ownerCheck(
	role *rbacv1beta1.Role,
	cluster *v1alpha1.ElasticsearchCluster,
) error {
	if !metav1.IsControlledBy(role, cluster) {
		ownerRef := metav1.GetControllerOf(role)
		return fmt.Errorf(
			"foreign owned Role: "+
				"A Role with name '%s/%s' already exists, "+
				"but it is controlled by '%v', not '%s/%s'",
			role.Namespace, role.Name, ownerRef,
			cluster.Namespace, cluster.Name,
		)
	}
	return nil
}

func (e *roleControl) Sync(c *v1alpha1.ElasticsearchCluster) error {
	desiredRole := roleForCluster(c)
	existingRole, err := e.roleLister.Roles(desiredRole.Namespace).Get(desiredRole.Name)
	if k8sErrors.IsNotFound(err) {
		_, err := e.kubeClient.RbacV1beta1().Roles(desiredRole.Namespace).Create(desiredRole)
		return err
	}
	if err != nil {
		return err
	}
	if err := ownerCheck(existingRole, c); err != nil {
		return err
	}
	// TODO: avoid using `reflect` here
	if !reflect.DeepEqual(desiredRole.Rules, existingRole.Rules) {
		_, err = e.kubeClient.RbacV1beta1().Roles(desiredRole.Namespace).Update(desiredRole)
		return err
	}
	return nil
}
