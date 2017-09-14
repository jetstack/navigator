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
package v1alpha1

import (
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	scheme "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PilotsGetter has a method to return a PilotInterface.
// A group's client should implement this interface.
type PilotsGetter interface {
	Pilots(namespace string) PilotInterface
}

// PilotInterface has methods to work with Pilot resources.
type PilotInterface interface {
	Create(*v1alpha1.Pilot) (*v1alpha1.Pilot, error)
	Update(*v1alpha1.Pilot) (*v1alpha1.Pilot, error)
	UpdateStatus(*v1alpha1.Pilot) (*v1alpha1.Pilot, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Pilot, error)
	List(opts v1.ListOptions) (*v1alpha1.PilotList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Pilot, err error)
	PilotExpansion
}

// pilots implements PilotInterface
type pilots struct {
	client rest.Interface
	ns     string
}

// newPilots returns a Pilots
func newPilots(c *NavigatorV1alpha1Client, namespace string) *pilots {
	return &pilots{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the pilot, and returns the corresponding pilot object, and an error if there is any.
func (c *pilots) Get(name string, options v1.GetOptions) (result *v1alpha1.Pilot, err error) {
	result = &v1alpha1.Pilot{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("pilots").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Pilots that match those selectors.
func (c *pilots) List(opts v1.ListOptions) (result *v1alpha1.PilotList, err error) {
	result = &v1alpha1.PilotList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("pilots").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested pilots.
func (c *pilots) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("pilots").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a pilot and creates it.  Returns the server's representation of the pilot, and an error, if there is any.
func (c *pilots) Create(pilot *v1alpha1.Pilot) (result *v1alpha1.Pilot, err error) {
	result = &v1alpha1.Pilot{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("pilots").
		Body(pilot).
		Do().
		Into(result)
	return
}

// Update takes the representation of a pilot and updates it. Returns the server's representation of the pilot, and an error, if there is any.
func (c *pilots) Update(pilot *v1alpha1.Pilot) (result *v1alpha1.Pilot, err error) {
	result = &v1alpha1.Pilot{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("pilots").
		Name(pilot.Name).
		Body(pilot).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *pilots) UpdateStatus(pilot *v1alpha1.Pilot) (result *v1alpha1.Pilot, err error) {
	result = &v1alpha1.Pilot{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("pilots").
		Name(pilot.Name).
		SubResource("status").
		Body(pilot).
		Do().
		Into(result)
	return
}

// Delete takes name of the pilot and deletes it. Returns an error if one occurs.
func (c *pilots) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("pilots").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *pilots) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("pilots").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched pilot.
func (c *pilots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Pilot, err error) {
	result = &v1alpha1.Pilot{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("pilots").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
