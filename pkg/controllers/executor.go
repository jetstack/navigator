package controllers

import (
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listers "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
)

// State contains the current state of the world, including accessors for
// modifying this state (e.g. kubernetes clientsets)
type State struct {
	// The Clientset to use when performing updates
	Clientset kubernetes.Interface
	// The NavigatorClientset to use for updates
	NavigatorClientset clientset.Interface
	Recorder           record.EventRecorder

	StatefulSetLister appslisters.StatefulSetLister
	ConfigMapLister   corelisters.ConfigMapLister
	PilotLister       listers.PilotLister
	PodLister         corelisters.PodLister
	ServiceLister     corelisters.ServiceLister
}

type Action interface {
	Name() string
	// Execute should attempt to execute the action. If it is not possible to
	// apply the specified changes (e.g. due to the cluster not being in a
	// 'ready state', or some transient error) then an error will be returned
	// so the action can be requeued. This allows for non-blocking blocking of
	// actions, with retries. The workqueues default scheduling and rate limit
	// will thus handle fairness within Navigator, and handle backing off on
	// retries.
	Execute(state *State) error
}
