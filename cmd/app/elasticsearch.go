package app

import (
	"gitlab.jetstack.net/marshal/colonel/pkg/elastic"
)

func newElasticsearchController(ctx *ControllerContext) (bool, error) {
	go elastic.NewElasticsearch(
		ctx.MarshalInformerFactory.V1().ElasticsearchCluster(),
		ctx.InformerFactory.Extensions().V1beta1().Deployments(),
		ctx.InformerFactory.Apps().V1beta1().StatefulSets(),
		ctx.Client,
	).Run(ctx.Stop)

	return true, nil
}
