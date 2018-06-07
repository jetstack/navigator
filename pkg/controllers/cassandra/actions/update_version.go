package actions

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
)

type UpdateVersion struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &UpdateVersion{}

func (a *UpdateVersion) Name() string {
	return "UpdateVersion"
}

func (a *UpdateVersion) Execute(s *controllers.State) error {
	baseSet := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	existingSet, err := s.StatefulSetLister.StatefulSets(baseSet.Namespace).Get(baseSet.Name)
	if err != nil {
		return errors.Wrap(err, "unable to get statefulset")
	}
	newImage := baseSet.Spec.Template.Spec.Containers[0].Image
	oldImage := existingSet.Spec.Template.Spec.Containers[0].Image
	if newImage == oldImage {
		glog.V(4).Infof(
			"StatefulSet %q already has the desired image %q",
			existingSet.Name, newImage,
		)
		return nil
	}
	glog.V(4).Infof(
		"Replacing StatefulSet %q image %q with %q",
		existingSet.Name, oldImage, newImage,
	)
	newSet := existingSet.DeepCopy()
	newSet.Spec.Template.Spec.Containers[0].Image = newImage
	_, err = s.Clientset.AppsV1beta1().StatefulSets(newSet.Namespace).Update(newSet)
	if err != nil {
		return errors.Wrap(err, "unable to update statefulset")
	}
	s.Recorder.Eventf(
		a.Cluster,
		corev1.EventTypeNormal,
		a.Name(),
		"UpdateVersion: NodePool=%q, Version=%q, Image=%q",
		a.NodePool.Name, a.Cluster.Spec.Version, newImage,
	)
	return nil
}
