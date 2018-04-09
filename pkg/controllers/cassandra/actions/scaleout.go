package actions

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/util/ptr"
)

type ScaleOut struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &ScaleOut{}

func (a *ScaleOut) Name() string {
	return "ScaleOut"
}

func (a *ScaleOut) Execute(s *controllers.State) error {
	baseSet := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	existingSet, err := s.StatefulSetLister.
		StatefulSets(baseSet.Namespace).Get(baseSet.Name)
	if err != nil {
		return errors.Wrap(err, "unable to find statefulset")
	}

	if *existingSet.Spec.Replicas == a.NodePool.Replicas {
		glog.V(4).Infof(
			"ScaleOut not necessary because StatefulSet '%s/%s' "+
				"already has the desired replicas value: %d",
			existingSet.Namespace, existingSet.Name,
			existingSet.Spec.Replicas,
		)
		return nil
	}

	if *existingSet.Spec.Replicas > a.NodePool.Replicas {
		glog.Errorf(
			"ScaleOut error. "+
				"The StatefulSet '%s/%s' replicas value must be lower than the desired value. "+
				"ActualReplicas: %d, DesiredReplicas: %d",
			existingSet.Namespace, existingSet.Name,
			*existingSet.Spec.Replicas, a.NodePool.Replicas,
		)
		return nil
	}

	if existingSet.Status.ReadyReplicas != *existingSet.Spec.Replicas {
		glog.V(4).Infof(
			"ScaleOut not possible because some pods in StatefulSet '%s/%s' are not ready. "+
				"DesiredReplicas: %d, CurrentReplicas: %d, ReadyReplicas: %d",
			existingSet.Namespace, existingSet.Name,
			*existingSet.Spec.Replicas,
			existingSet.Status.CurrentReplicas, existingSet.Status.ReadyReplicas,
		)
		return nil
	}

	newSet := existingSet.DeepCopy()
	newSet.Spec.Replicas = ptr.Int32(*newSet.Spec.Replicas + 1)
	glog.V(4).Infof(
		"Scaling statefulset %s/%s from %d to %d",
		newSet.Namespace, newSet.Name,
		existingSet.Spec.Replicas, newSet.Spec.Replicas,
	)
	_, err = s.Clientset.AppsV1beta1().
		StatefulSets(newSet.Namespace).Update(newSet)
	if err != nil {
		return errors.Wrap(err, "unable to update statefulset replica count")
	}
	s.Recorder.Eventf(
		a.Cluster,
		corev1.EventTypeNormal,
		a.Name(),
		"ScaleOut: NodePool=%s/%s/%s, ReplicaCount=%d, TargetReplicaCount=%d",
		a.Cluster.Namespace, a.Cluster.Name, a.NodePool.Name,
		*existingSet.Spec.Replicas, a.NodePool.Replicas,
	)
	return nil
}
