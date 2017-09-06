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

// FakeElasticsearchClusters implements ElasticsearchClusterInterface
type FakeElasticsearchClusters struct {
	Fake *FakeNavigator
	ns   string
}

var elasticsearchclustersResource = schema.GroupVersionResource{Group: "navigator.jetstack.io", Version: "", Resource: "elasticsearchclusters"}

var elasticsearchclustersKind = schema.GroupVersionKind{Group: "navigator.jetstack.io", Version: "", Kind: "ElasticsearchCluster"}

// Get takes name of the elasticsearchCluster, and returns the corresponding elasticsearchCluster object, and an error if there is any.
func (c *FakeElasticsearchClusters) Get(name string, options v1.GetOptions) (result *navigator.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(elasticsearchclustersResource, c.ns, name), &navigator.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.ElasticsearchCluster), err
}

// List takes label and field selectors, and returns the list of ElasticsearchClusters that match those selectors.
func (c *FakeElasticsearchClusters) List(opts v1.ListOptions) (result *navigator.ElasticsearchClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(elasticsearchclustersResource, elasticsearchclustersKind, c.ns, opts), &navigator.ElasticsearchClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &navigator.ElasticsearchClusterList{}
	for _, item := range obj.(*navigator.ElasticsearchClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested elasticsearchClusters.
func (c *FakeElasticsearchClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(elasticsearchclustersResource, c.ns, opts))

}

// Create takes the representation of a elasticsearchCluster and creates it.  Returns the server's representation of the elasticsearchCluster, and an error, if there is any.
func (c *FakeElasticsearchClusters) Create(elasticsearchCluster *navigator.ElasticsearchCluster) (result *navigator.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(elasticsearchclustersResource, c.ns, elasticsearchCluster), &navigator.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.ElasticsearchCluster), err
}

// Update takes the representation of a elasticsearchCluster and updates it. Returns the server's representation of the elasticsearchCluster, and an error, if there is any.
func (c *FakeElasticsearchClusters) Update(elasticsearchCluster *navigator.ElasticsearchCluster) (result *navigator.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(elasticsearchclustersResource, c.ns, elasticsearchCluster), &navigator.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.ElasticsearchCluster), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeElasticsearchClusters) UpdateStatus(elasticsearchCluster *navigator.ElasticsearchCluster) (*navigator.ElasticsearchCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(elasticsearchclustersResource, "status", c.ns, elasticsearchCluster), &navigator.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.ElasticsearchCluster), err
}

// Delete takes name of the elasticsearchCluster and deletes it. Returns an error if one occurs.
func (c *FakeElasticsearchClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(elasticsearchclustersResource, c.ns, name), &navigator.ElasticsearchCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeElasticsearchClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(elasticsearchclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &navigator.ElasticsearchClusterList{})
	return err
}

// Patch applies the patch and returns the patched elasticsearchCluster.
func (c *FakeElasticsearchClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *navigator.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(elasticsearchclustersResource, c.ns, name, data, subresources...), &navigator.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.ElasticsearchCluster), err
}
