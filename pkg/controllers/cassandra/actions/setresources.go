package actions

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/util/resources"
)

type SetResources struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &SetResources{}

func (a *SetResources) Name() string {
	return "SetResources"
}

func (a *SetResources) Execute(s *controllers.State) error {
	baseSet := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	existingSet, err := s.StatefulSetLister.
		StatefulSets(baseSet.Namespace).Get(baseSet.Name)
	if err != nil {
		return errors.Wrap(err, "unable to find statefulset")
	}

	var cassContainerIndex int
	var container *corev1.Container
	for i, _ := range existingSet.Spec.Template.Spec.Containers {
		if existingSet.Spec.Template.Spec.Containers[i].Name == "cassandra" {
			cassContainerIndex = i
			container = &existingSet.Spec.Template.Spec.Containers[i]
		}
	}

	if container == nil {
		return fmt.Errorf("unable to find cassandra container in StatefulSet %s/%s",
			existingSet.Namespace, existingSet.Name,
		)
	}

	if resources.RequirementsEqual(container.Resources, a.NodePool.Resources) {
		glog.V(4).Infof(
			"SetResources not necessary because StatefulSet '%s/%s' "+
				"already has the desired resources value: %v",
			existingSet.Namespace, existingSet.Name,
			container.Resources,
		)
		return nil
	}

	newSet := existingSet.DeepCopy()
	newSet.Spec.Template.Spec.Containers[cassContainerIndex].Resources = a.NodePool.Resources
	glog.V(4).Infof(
		"Setting cassandra resources %s/%s from %v to %v",
		newSet.Namespace, newSet.Name,
		existingSet.Spec.Template.Spec.Containers[cassContainerIndex].Resources,
		a.NodePool.Resources,
	)
	_, err = s.Clientset.AppsV1beta1().
		StatefulSets(newSet.Namespace).Update(newSet)
	if err != nil {
		return errors.Wrap(err, "unable to update statefulset resources")
	}
	s.Recorder.Eventf(
		a.Cluster,
		corev1.EventTypeNormal,
		a.Name(),
		"SetResources: NodePool=%s/%s/%s, Resources=%v",
		a.Cluster.Namespace, a.Cluster.Name, a.NodePool.Name,
		a.NodePool.Resources,
	)
	return nil
}
