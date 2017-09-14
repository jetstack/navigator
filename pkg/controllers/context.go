package controllers

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	"github.com/jetstack-experimental/navigator/pkg/kube"
)

type Context struct {
	Client          kubernetes.Interface
	NavigatorClient clientset.Interface

	// Recorder to record events to
	Recorder              record.EventRecorder
	SharedInformerFactory kube.SharedInformerFactory

	Namespace string
}
