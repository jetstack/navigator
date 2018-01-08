/*
Copyright 2018 Jetstack Ltd.

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
package internalversion

import (
	navigator "github.com/jetstack/navigator/pkg/apis/navigator"
	scheme "github.com/jetstack/navigator/pkg/client/clientset/internalversion/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CassandraClustersGetter has a method to return a CassandraClusterInterface.
// A group's client should implement this interface.
type CassandraClustersGetter interface {
	CassandraClusters(namespace string) CassandraClusterInterface
}

// CassandraClusterInterface has methods to work with CassandraCluster resources.
type CassandraClusterInterface interface {
	Create(*navigator.CassandraCluster) (*navigator.CassandraCluster, error)
	Update(*navigator.CassandraCluster) (*navigator.CassandraCluster, error)
	UpdateStatus(*navigator.CassandraCluster) (*navigator.CassandraCluster, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*navigator.CassandraCluster, error)
	List(opts v1.ListOptions) (*navigator.CassandraClusterList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *navigator.CassandraCluster, err error)
	CassandraClusterExpansion
}

// cassandraClusters implements CassandraClusterInterface
type cassandraClusters struct {
	client rest.Interface
	ns     string
}

// newCassandraClusters returns a CassandraClusters
func newCassandraClusters(c *NavigatorClient, namespace string) *cassandraClusters {
	return &cassandraClusters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cassandraCluster, and returns the corresponding cassandraCluster object, and an error if there is any.
func (c *cassandraClusters) Get(name string, options v1.GetOptions) (result *navigator.CassandraCluster, err error) {
	result = &navigator.CassandraCluster{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cassandraclusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CassandraClusters that match those selectors.
func (c *cassandraClusters) List(opts v1.ListOptions) (result *navigator.CassandraClusterList, err error) {
	result = &navigator.CassandraClusterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cassandraclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cassandraClusters.
func (c *cassandraClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cassandraclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cassandraCluster and creates it.  Returns the server's representation of the cassandraCluster, and an error, if there is any.
func (c *cassandraClusters) Create(cassandraCluster *navigator.CassandraCluster) (result *navigator.CassandraCluster, err error) {
	result = &navigator.CassandraCluster{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cassandraclusters").
		Body(cassandraCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cassandraCluster and updates it. Returns the server's representation of the cassandraCluster, and an error, if there is any.
func (c *cassandraClusters) Update(cassandraCluster *navigator.CassandraCluster) (result *navigator.CassandraCluster, err error) {
	result = &navigator.CassandraCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cassandraclusters").
		Name(cassandraCluster.Name).
		Body(cassandraCluster).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *cassandraClusters) UpdateStatus(cassandraCluster *navigator.CassandraCluster) (result *navigator.CassandraCluster, err error) {
	result = &navigator.CassandraCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cassandraclusters").
		Name(cassandraCluster.Name).
		SubResource("status").
		Body(cassandraCluster).
		Do().
		Into(result)
	return
}

// Delete takes name of the cassandraCluster and deletes it. Returns an error if one occurs.
func (c *cassandraClusters) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cassandraclusters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cassandraClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cassandraclusters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cassandraCluster.
func (c *cassandraClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *navigator.CassandraCluster, err error) {
	result = &navigator.CassandraCluster{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cassandraclusters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
