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
		return errors.Wrap(err, "unable to get existing statefulset")
	}
	if *existingSet.Spec.Replicas == a.NodePool.Replicas {
		glog.V(4).Infof(
			"The StatefulSet %s/%s already has the desired replicas value %d",
			existingSet.Namespace, existingSet.Name, a.NodePool.Replicas,
		)
		return nil
	}
	if *existingSet.Spec.Replicas > a.NodePool.Replicas {
		glog.Errorf(
			"ScaleOut error:"+
				"The StatefulSet %s/%s replicas value (%d) "+
				"is greater than the desired value (%d)",
			existingSet.Namespace, existingSet.Name,
			*existingSet.Spec.Replicas, a.NodePool.Replicas,
		)
		return nil
	}
	if *existingSet.Spec.Replicas != existingSet.Status.ReadyReplicas {
		glog.V(4).Infof(
			"Waiting for all pods in nodepool %s/%s/%s (statefulset %s/%s) to become ready "+
				"before adding more nodes. "+
				"Replicas: %d, ReadyReplicas: %d",
			a.Cluster.Namespace, a.Cluster.Name, a.NodePool.Name,
			existingSet.Namespace, existingSet.Name,
			*existingSet.Spec.Replicas, existingSet.Status.ReadyReplicas,
		)
		return nil
	}
	newSet := existingSet.DeepCopy()
	newSet.Spec.Replicas = ptr.Int32(*newSet.Spec.Replicas + 1)
	_, err = s.Clientset.AppsV1beta1().
		StatefulSets(newSet.Namespace).Update(newSet)
	if err != nil {
		return errors.Wrap(err, "unable to update statefulset")
	}
	s.Recorder.Eventf(
		a.Cluster,
		corev1.EventTypeNormal,
		a.Name(),
		"ScaleOut: Cluster=%s/%s, NodePool=%q, ReplicaCount=%d, TargetReplicaCount=%d",
		a.Cluster.Namespace, a.Cluster.Name,
		a.NodePool.Name, *newSet.Spec.Replicas, a.NodePool.Replicas,
	)
	return nil
}
