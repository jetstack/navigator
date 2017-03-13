package app

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	intinformers "gitlab.jetstack.net/marshal/colonel/pkg/informers"
)

var (
	known = map[string]InitFn{
		"ElasticSearch": newElasticsearchController,
	}
)

func Known() map[string]InitFn {
	return known
}

type ControllerContext struct {
	Client    *kubernetes.Clientset
	TPRClient *rest.RESTClient

	InformerFactory        informers.SharedInformerFactory
	MarshalInformerFactory intinformers.SharedInformerFactory

	Namespace string
	Stop      <-chan struct{}
}

type InitFn func(*ControllerContext) (bool, error)

func StartControllers(ctx *ControllerContext, fns map[string]InitFn, stop <-chan struct{}) error {
	for n, fn := range fns {
		logrus.Debugf("starting %s controller", n)

		_, err := fn(ctx)

		if err != nil {
			return fmt.Errorf("error starting '%s' controller: %s", n, err.Error())
		}
	}

	ctx.InformerFactory.Start(stop)
	ctx.MarshalInformerFactory.Start(stop)

	select {}
}
