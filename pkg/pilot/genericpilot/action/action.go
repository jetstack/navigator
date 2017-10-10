package action

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

type Interface interface {
	// The name of the Action to perform (e.g. Decommission)
	Name() string
	// Execute the action with the given Pilot resource
	Execute(pilot *v1alpha1.Pilot) error
}

func New(name string, fn func(*v1alpha1.Pilot) error) Interface {
	return &actionAdapter{
		name: name,
		fn:   fn,
	}
}

type actionAdapter struct {
	name string
	fn   func(pilot *v1alpha1.Pilot) error
}

var _ Interface = &actionAdapter{}

func (h *actionAdapter) Name() string {
	return h.name
}

func (h *actionAdapter) Execute(pilot *v1alpha1.Pilot) error {
	return h.fn(pilot)
}
