/*
Copyright 2017 Jetstack Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package clientset

import (
	glog "github.com/golang/glog"
	navigatorv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/typed/navigator/v1alpha1"
	navigatorv1alpha2 "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/typed/navigator/v1alpha2"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	NavigatorV1alpha1() navigatorv1alpha1.NavigatorV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Navigator() navigatorv1alpha1.NavigatorV1alpha1Interface
	NavigatorV1alpha2() navigatorv1alpha2.NavigatorV1alpha2Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	navigatorV1alpha1 *navigatorv1alpha1.NavigatorV1alpha1Client
	navigatorV1alpha2 *navigatorv1alpha2.NavigatorV1alpha2Client
}

// NavigatorV1alpha1 retrieves the NavigatorV1alpha1Client
func (c *Clientset) NavigatorV1alpha1() navigatorv1alpha1.NavigatorV1alpha1Interface {
	return c.navigatorV1alpha1
}

// Deprecated: Navigator retrieves the default version of NavigatorClient.
// Please explicitly pick a version.
func (c *Clientset) Navigator() navigatorv1alpha1.NavigatorV1alpha1Interface {
	return c.navigatorV1alpha1
}

// NavigatorV1alpha2 retrieves the NavigatorV1alpha2Client
func (c *Clientset) NavigatorV1alpha2() navigatorv1alpha2.NavigatorV1alpha2Interface {
	return c.navigatorV1alpha2
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.navigatorV1alpha1, err = navigatorv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.navigatorV1alpha2, err = navigatorv1alpha2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.navigatorV1alpha1 = navigatorv1alpha1.NewForConfigOrDie(c)
	cs.navigatorV1alpha2 = navigatorv1alpha2.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.navigatorV1alpha1 = navigatorv1alpha1.New(c)
	cs.navigatorV1alpha2 = navigatorv1alpha2.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
