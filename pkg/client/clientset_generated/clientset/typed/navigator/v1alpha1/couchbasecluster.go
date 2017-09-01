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

// CouchbaseClustersGetter has a method to return a CouchbaseClusterInterface.
// A group's client should implement this interface.
type CouchbaseClustersGetter interface {
	CouchbaseClusters(namespace string) CouchbaseClusterInterface
}

// CouchbaseClusterInterface has methods to work with CouchbaseCluster resources.
type CouchbaseClusterInterface interface {
	Create(*v1alpha1.CouchbaseCluster) (*v1alpha1.CouchbaseCluster, error)
	Update(*v1alpha1.CouchbaseCluster) (*v1alpha1.CouchbaseCluster, error)
	UpdateStatus(*v1alpha1.CouchbaseCluster) (*v1alpha1.CouchbaseCluster, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CouchbaseCluster, error)
	List(opts v1.ListOptions) (*v1alpha1.CouchbaseClusterList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CouchbaseCluster, err error)
	CouchbaseClusterExpansion
}

// couchbaseClusters implements CouchbaseClusterInterface
type couchbaseClusters struct {
	client rest.Interface
	ns     string
}

// newCouchbaseClusters returns a CouchbaseClusters
func newCouchbaseClusters(c *NavigatorV1alpha1Client, namespace string) *couchbaseClusters {
	return &couchbaseClusters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the couchbaseCluster, and returns the corresponding couchbaseCluster object, and an error if there is any.
func (c *couchbaseClusters) Get(name string, options v1.GetOptions) (result *v1alpha1.CouchbaseCluster, err error) {
	result = &v1alpha1.CouchbaseCluster{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CouchbaseClusters that match those selectors.
func (c *couchbaseClusters) List(opts v1.ListOptions) (result *v1alpha1.CouchbaseClusterList, err error) {
	result = &v1alpha1.CouchbaseClusterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested couchbaseClusters.
func (c *couchbaseClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a couchbaseCluster and creates it.  Returns the server's representation of the couchbaseCluster, and an error, if there is any.
func (c *couchbaseClusters) Create(couchbaseCluster *v1alpha1.CouchbaseCluster) (result *v1alpha1.CouchbaseCluster, err error) {
	result = &v1alpha1.CouchbaseCluster{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		Body(couchbaseCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a couchbaseCluster and updates it. Returns the server's representation of the couchbaseCluster, and an error, if there is any.
func (c *couchbaseClusters) Update(couchbaseCluster *v1alpha1.CouchbaseCluster) (result *v1alpha1.CouchbaseCluster, err error) {
	result = &v1alpha1.CouchbaseCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		Name(couchbaseCluster.Name).
		Body(couchbaseCluster).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *couchbaseClusters) UpdateStatus(couchbaseCluster *v1alpha1.CouchbaseCluster) (result *v1alpha1.CouchbaseCluster, err error) {
	result = &v1alpha1.CouchbaseCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		Name(couchbaseCluster.Name).
		SubResource("status").
		Body(couchbaseCluster).
		Do().
		Into(result)
	return
}

// Delete takes name of the couchbaseCluster and deletes it. Returns an error if one occurs.
func (c *couchbaseClusters) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *couchbaseClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("couchbaseclusters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched couchbaseCluster.
func (c *couchbaseClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CouchbaseCluster, err error) {
	result = &v1alpha1.CouchbaseCluster{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("couchbaseclusters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
