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
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCassandraClusters implements CassandraClusterInterface
type FakeCassandraClusters struct {
	Fake *FakeNavigatorV1alpha1
	ns   string
}

var cassandraclustersResource = schema.GroupVersionResource{Group: "navigator.jetstack.io", Version: "v1alpha1", Resource: "cassandraclusters"}

var cassandraclustersKind = schema.GroupVersionKind{Group: "navigator.jetstack.io", Version: "v1alpha1", Kind: "CassandraCluster"}

// Get takes name of the cassandraCluster, and returns the corresponding cassandraCluster object, and an error if there is any.
func (c *FakeCassandraClusters) Get(name string, options v1.GetOptions) (result *v1alpha1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cassandraclustersResource, c.ns, name), &v1alpha1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CassandraCluster), err
}

// List takes label and field selectors, and returns the list of CassandraClusters that match those selectors.
func (c *FakeCassandraClusters) List(opts v1.ListOptions) (result *v1alpha1.CassandraClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cassandraclustersResource, cassandraclustersKind, c.ns, opts), &v1alpha1.CassandraClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CassandraClusterList{}
	for _, item := range obj.(*v1alpha1.CassandraClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cassandraClusters.
func (c *FakeCassandraClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cassandraclustersResource, c.ns, opts))

}

// Create takes the representation of a cassandraCluster and creates it.  Returns the server's representation of the cassandraCluster, and an error, if there is any.
func (c *FakeCassandraClusters) Create(cassandraCluster *v1alpha1.CassandraCluster) (result *v1alpha1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cassandraclustersResource, c.ns, cassandraCluster), &v1alpha1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CassandraCluster), err
}

// Update takes the representation of a cassandraCluster and updates it. Returns the server's representation of the cassandraCluster, and an error, if there is any.
func (c *FakeCassandraClusters) Update(cassandraCluster *v1alpha1.CassandraCluster) (result *v1alpha1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cassandraclustersResource, c.ns, cassandraCluster), &v1alpha1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CassandraCluster), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCassandraClusters) UpdateStatus(cassandraCluster *v1alpha1.CassandraCluster) (*v1alpha1.CassandraCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(cassandraclustersResource, "status", c.ns, cassandraCluster), &v1alpha1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CassandraCluster), err
}

// Delete takes name of the cassandraCluster and deletes it. Returns an error if one occurs.
func (c *FakeCassandraClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cassandraclustersResource, c.ns, name), &v1alpha1.CassandraCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCassandraClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cassandraclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CassandraClusterList{})
	return err
}

// Patch applies the patch and returns the patched cassandraCluster.
func (c *FakeCassandraClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CassandraCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cassandraclustersResource, c.ns, name, data, subresources...), &v1alpha1.CassandraCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CassandraCluster), err
}
