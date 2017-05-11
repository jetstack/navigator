package controllers

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	intinformers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
)

type Context struct {
	Client *kubernetes.Clientset

	InformerFactory        informers.SharedInformerFactory
	MarshalInformerFactory intinformers.SharedInformerFactory

	Namespace string
	Stop      <-chan struct{}
}

type InitFn func(*Context) (bool, error)

func Start(ctx *Context, fns map[string]InitFn, stop <-chan struct{}) error {
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
