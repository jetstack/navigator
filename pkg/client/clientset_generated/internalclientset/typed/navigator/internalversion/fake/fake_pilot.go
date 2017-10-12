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

// FakePilots implements PilotInterface
type FakePilots struct {
	Fake *FakeNavigator
	ns   string
}

var pilotsResource = schema.GroupVersionResource{Group: "navigator.jetstack.io", Version: "", Resource: "pilots"}

var pilotsKind = schema.GroupVersionKind{Group: "navigator.jetstack.io", Version: "", Kind: "Pilot"}

// Get takes name of the pilot, and returns the corresponding pilot object, and an error if there is any.
func (c *FakePilots) Get(name string, options v1.GetOptions) (result *navigator.Pilot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(pilotsResource, c.ns, name), &navigator.Pilot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.Pilot), err
}

// List takes label and field selectors, and returns the list of Pilots that match those selectors.
func (c *FakePilots) List(opts v1.ListOptions) (result *navigator.PilotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(pilotsResource, pilotsKind, c.ns, opts), &navigator.PilotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &navigator.PilotList{}
	for _, item := range obj.(*navigator.PilotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pilots.
func (c *FakePilots) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(pilotsResource, c.ns, opts))

}

// Create takes the representation of a pilot and creates it.  Returns the server's representation of the pilot, and an error, if there is any.
func (c *FakePilots) Create(pilot *navigator.Pilot) (result *navigator.Pilot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(pilotsResource, c.ns, pilot), &navigator.Pilot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.Pilot), err
}

// Update takes the representation of a pilot and updates it. Returns the server's representation of the pilot, and an error, if there is any.
func (c *FakePilots) Update(pilot *navigator.Pilot) (result *navigator.Pilot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(pilotsResource, c.ns, pilot), &navigator.Pilot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.Pilot), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePilots) UpdateStatus(pilot *navigator.Pilot) (*navigator.Pilot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(pilotsResource, "status", c.ns, pilot), &navigator.Pilot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.Pilot), err
}

// Delete takes name of the pilot and deletes it. Returns an error if one occurs.
func (c *FakePilots) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(pilotsResource, c.ns, name), &navigator.Pilot{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePilots) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(pilotsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &navigator.PilotList{})
	return err
}

// Patch applies the patch and returns the patched pilot.
func (c *FakePilots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *navigator.Pilot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(pilotsResource, c.ns, name, data, subresources...), &navigator.Pilot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*navigator.Pilot), err
}
