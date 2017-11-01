package controllers

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	intinformers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	kubeinformers "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions"
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
