package controllers

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	"github.com/jetstack-experimental/navigator/pkg/kube"
)

type Context struct {
	Client          kubernetes.Interface
	NavigatorClient clientset.Interface

	SharedInformerFactory kube.SharedInformerFactory

	Namespace string
}

type InitFn func(ctx *Context, stopCh <-chan struct{}) (bool, error)

func Start(ctx *Context, fns map[string]InitFn, stopCh <-chan struct{}) error {
	for n, fn := range fns {
		logrus.Debugf("starting %s controller", n)

		_, err := fn(ctx, stopCh)

		if err != nil {
			return fmt.Errorf("error starting '%s' controller: %s", n, err.Error())
		}
	}

	ctx.SharedInformerFactory.Start(stopCh)

	select {}
}
