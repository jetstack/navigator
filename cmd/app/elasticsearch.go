package app

import "gitlab.jetstack.net/marshal/colonel/pkg/controllers/elasticsearch"

func newElasticsearchController(ctx *ControllerContext) (bool, error) {
	go elasticsearch.NewElasticsearch(
		ctx.MarshalInformerFactory.V1().ElasticsearchCluster(),
		ctx.InformerFactory.Extensions().V1beta1().Deployments(),
		ctx.InformerFactory.Apps().V1beta1().StatefulSets(),
		ctx.Client,
	).Run(2, ctx.Stop)

	return true, nil
}
