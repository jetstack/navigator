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
package fake

import (
	navigator "github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCouchbaseClusters implements CouchbaseClusterInterface
type FakeCouchbaseClusters struct {
	Fake *FakeNavigator
	ns   string
}

var couchbaseclustersResource = schema.GroupVersionResource{Group: "navigator.jetstack.io", Version: "", Resource: "couchbaseclusters"}

var couchbaseclustersKind = schema.GroupVersionKind{Group: "navigator.jetstack.io", Version: "", Kind: "CouchbaseCluster"}

// Get takes name of the couchbaseCluster, and returns the corresponding couchbaseCluster object, and an error if there is any.
func (c *FakeCouchbaseClusters) Get(name string, options v1.GetOptions) (result *navigator.CouchbaseCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(couchbaseclustersResource, c.ns, name), &navigator.CouchbaseCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.CouchbaseCluster), err
}

// List takes label and field selectors, and returns the list of CouchbaseClusters that match those selectors.
func (c *FakeCouchbaseClusters) List(opts v1.ListOptions) (result *navigator.CouchbaseClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(couchbaseclustersResource, couchbaseclustersKind, c.ns, opts), &navigator.CouchbaseClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &navigator.CouchbaseClusterList{}
	for _, item := range obj.(*navigator.CouchbaseClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested couchbaseClusters.
func (c *FakeCouchbaseClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(couchbaseclustersResource, c.ns, opts))

}

// Create takes the representation of a couchbaseCluster and creates it.  Returns the server's representation of the couchbaseCluster, and an error, if there is any.
func (c *FakeCouchbaseClusters) Create(couchbaseCluster *navigator.CouchbaseCluster) (result *navigator.CouchbaseCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(couchbaseclustersResource, c.ns, couchbaseCluster), &navigator.CouchbaseCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.CouchbaseCluster), err
}

// Update takes the representation of a couchbaseCluster and updates it. Returns the server's representation of the couchbaseCluster, and an error, if there is any.
func (c *FakeCouchbaseClusters) Update(couchbaseCluster *navigator.CouchbaseCluster) (result *navigator.CouchbaseCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(couchbaseclustersResource, c.ns, couchbaseCluster), &navigator.CouchbaseCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.CouchbaseCluster), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCouchbaseClusters) UpdateStatus(couchbaseCluster *navigator.CouchbaseCluster) (*navigator.CouchbaseCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(couchbaseclustersResource, "status", c.ns, couchbaseCluster), &navigator.CouchbaseCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.CouchbaseCluster), err
}

// Delete takes name of the couchbaseCluster and deletes it. Returns an error if one occurs.
func (c *FakeCouchbaseClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(couchbaseclustersResource, c.ns, name), &navigator.CouchbaseCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCouchbaseClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(couchbaseclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &navigator.CouchbaseClusterList{})
	return err
}

// Patch applies the patch and returns the patched couchbaseCluster.
func (c *FakeCouchbaseClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *navigator.CouchbaseCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(couchbaseclustersResource, c.ns, name, data, subresources...), &navigator.CouchbaseCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.CouchbaseCluster), err
}
