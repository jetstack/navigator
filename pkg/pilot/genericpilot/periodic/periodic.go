package periodic

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

type Interface interface {
	// The name of the Periodic
	Name() string
	// Execute the periodic with the given Pilot resource
	Execute(pilot *v1alpha1.Pilot) error
}

func New(name string, fn func(*v1alpha1.Pilot) error) Interface {
	return &periodicAdapter{
		name: name,
		fn:   fn,
	}
}

type periodicAdapter struct {
	name string
	fn   func(pilot *v1alpha1.Pilot) error
}

var _ Interface = &periodicAdapter{}

func (h *periodicAdapter) Name() string {
	return h.name
}

func (h *periodicAdapter) Execute(pilot *v1alpha1.Pilot) error {
	return h.fn(pilot)
}
