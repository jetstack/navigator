/*
Copyright 2017 The Kubernetes Authors.

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
	marshal "github.com/jetstack-experimental/navigator/pkg/apis/marshal"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeElasticsearchClusters implements ElasticsearchClusterInterface
type FakeElasticsearchClusters struct {
	Fake *FakeMarshal
	ns   string
}

var elasticsearchclustersResource = schema.GroupVersionResource{Group: "marshal.io", Version: "", Resource: "elasticsearchclusters"}

var elasticsearchclustersKind = schema.GroupVersionKind{Group: "marshal.io", Version: "", Kind: "ElasticsearchCluster"}

func (c *FakeElasticsearchClusters) Create(elasticsearchCluster *marshal.ElasticsearchCluster) (result *marshal.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(elasticsearchclustersResource, c.ns, elasticsearchCluster), &marshal.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*marshal.ElasticsearchCluster), err
}

func (c *FakeElasticsearchClusters) Update(elasticsearchCluster *marshal.ElasticsearchCluster) (result *marshal.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(elasticsearchclustersResource, c.ns, elasticsearchCluster), &marshal.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*marshal.ElasticsearchCluster), err
}

func (c *FakeElasticsearchClusters) UpdateStatus(elasticsearchCluster *marshal.ElasticsearchCluster) (*marshal.ElasticsearchCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(elasticsearchclustersResource, "status", c.ns, elasticsearchCluster), &marshal.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*marshal.ElasticsearchCluster), err
}

func (c *FakeElasticsearchClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(elasticsearchclustersResource, c.ns, name), &marshal.ElasticsearchCluster{})

	return err
}

func (c *FakeElasticsearchClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(elasticsearchclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &marshal.ElasticsearchClusterList{})
	return err
}

func (c *FakeElasticsearchClusters) Get(name string, options v1.GetOptions) (result *marshal.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(elasticsearchclustersResource, c.ns, name), &marshal.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*marshal.ElasticsearchCluster), err
}

func (c *FakeElasticsearchClusters) List(opts v1.ListOptions) (result *marshal.ElasticsearchClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(elasticsearchclustersResource, elasticsearchclustersKind, c.ns, opts), &marshal.ElasticsearchClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &marshal.ElasticsearchClusterList{}
	for _, item := range obj.(*marshal.ElasticsearchClusterList).Items {
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

// Patch applies the patch and returns the patched elasticsearchCluster.
func (c *FakeElasticsearchClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *marshal.ElasticsearchCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(elasticsearchclustersResource, c.ns, name, data, subresources...), &marshal.ElasticsearchCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*marshal.ElasticsearchCluster), err
}
