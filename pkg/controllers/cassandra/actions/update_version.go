package actions

import (
	"fmt"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type UpdateVersion struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &UpdateVersion{}

func (c *UpdateVersion) Name() string {
	return "UpdateVersion"
}

func (c *UpdateVersion) Execute(state *controllers.State) error {
	statefulSetName := util.NodePoolResourceName(c.Cluster, c.NodePool)
	statefulSet, err := state.StatefulSetLister.StatefulSets(c.Cluster.Namespace).Get(statefulSetName)
	if err != nil {
		return err
	}
	statefulSet = statefulSet.DeepCopy()
	newImage := nodepool.CassImageToUse(&c.Cluster.Spec)
	newImageString := fmt.Sprintf(
		"%s:%s",
		newImage.Repository,
		newImage.Tag,
	)
	oldImageString := statefulSet.Spec.Template.Spec.Containers[0].Image
	if newImageString == oldImageString {
		return nil
	}
	statefulSet.Spec.Template.Spec.Containers[0].Image = newImageString
	_, err = state.Clientset.AppsV1beta1().StatefulSets(statefulSet.Namespace).Update(statefulSet)
	return err
}
