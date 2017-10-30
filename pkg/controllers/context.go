package controllers

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	"github.com/jetstack/navigator/pkg/kube"
)

type Context struct {
	Client          kubernetes.Interface
	NavigatorClient clientset.Interface

	// Recorder to record events to
	Recorder              record.EventRecorder
	SharedInformerFactory kube.SharedInformerFactory

	Namespace string
}
