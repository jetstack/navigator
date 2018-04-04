package actions

import (
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pkg/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
)

type CreateNodePool struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &CreateNodePool{}

func (a *CreateNodePool) Name() string {
	return "CreateNodePool"
}

func (a *CreateNodePool) Execute(s *controllers.State) error {
	ss := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	_, err := s.Clientset.AppsV1beta1().StatefulSets(ss.Namespace).Create(ss)
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "unable to create statefulset")
	}
	s.Recorder.Eventf(
		a.Cluster,
		corev1.EventTypeNormal,
		a.Name(),
		"CreateNodePool: Name=%q", a.NodePool.Name,
	)
	return nil
}
