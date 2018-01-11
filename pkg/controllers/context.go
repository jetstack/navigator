package controllers

import (
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	intinformers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
)

type Context struct {
	Client          kubernetes.Interface
	NavigatorClient clientset.Interface

	// Recorder to record events to
	Recorder                  record.EventRecorder
	KubeSharedInformerFactory kubeinformers.SharedInformerFactory
	SharedInformerFactory     intinformers.SharedInformerFactory

	Namespace string
}
