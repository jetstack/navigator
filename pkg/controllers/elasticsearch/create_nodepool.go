package elasticsearch

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/nodepool"
)

type CreateNodePool struct {
	Cluster  *v1alpha1.ElasticsearchCluster
	NodePool *v1alpha1.ElasticsearchClusterNodePool
}

var _ controllers.Action = &CreateNodePool{}

func (c *CreateNodePool) Name() string {
	return "CreateNodePool"
}

func (c *CreateNodePool) Execute(state *controllers.State) error {
	toCreate, err := nodepool.NodePoolStatefulSet(c.Cluster, c.NodePool)
	if err != nil {
		return err
	}

	_, err = state.Clientset.AppsV1beta1().StatefulSets(toCreate.Namespace).Create(toCreate)
	if err != nil {
		return err
	}

	return nil
}
