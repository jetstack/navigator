package hook

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"sync"
)

type Interface interface {
	// The name of the Hook
	Name() string
	// Execute the action with the given Pilot resource
	Execute(pilot *v1alpha1.Pilot) error
}

func New(name string, fn func(*v1alpha1.Pilot) error) Interface {
	return &hookAdapter{
		name: name,
		fn:   fn,
	}
}

type hookAdapter struct {
	name string
	fn   func(pilot *v1alpha1.Pilot) error
}

var _ Interface = &hookAdapter{}

func (h *hookAdapter) Name() string {
	return h.name
}

func (h *hookAdapter) Execute(pilot *v1alpha1.Pilot) error {
	return h.fn(pilot)
}

type Hooks struct {
	// PreStart are hooks to be run before the application starts
	PreStart         []Interface
	executedPreStart map[string]struct{}
	// PostStart are hooks to be run after the application starts
	PostStart         []Interface
	executedPostStart map[string]struct{}
	// PreStop are hooks to be run before the application stops
	PreStop         []Interface
	executedPreStop map[string]struct{}
	// PostStop are hooks to be run after the application stops
	PostStop         []Interface
	executedPostStop map[string]struct{}

	lock sync.Mutex
}

type Phase string

const (
	PhasePreStart  Phase = "PreStart"
	PhasePostStart Phase = "PostStart"
	PhasePreStop   Phase = "PreStop"
	PhasePostStop  Phase = "PostStop"
)

func (h *Hooks) Transition(p Phase, pilot *v1alpha1.Pilot) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	if pilot == nil {
		return fmt.Errorf("pilot resource must not be nil")
	}
	var hooks []Interface
	executed := map[string]struct{}{}
	switch p {
	case PhasePreStart:
		hooks = h.PreStart
		if h.executedPreStart == nil {
			h.executedPreStart = executed
		}
		executed = h.executedPreStart
	case PhasePostStart:
		hooks = h.PostStart
		if h.executedPostStart == nil {
			h.executedPostStart = executed
		}
		executed = h.executedPostStart
	case PhasePreStop:
		hooks = h.PreStop
		if h.executedPreStop == nil {
			h.executedPreStop = executed
		}
		executed = h.executedPreStop
	case PhasePostStop:
		hooks = h.PostStop
		if h.executedPostStop == nil {
			h.executedPostStop = executed
		}
		executed = h.executedPostStop
	default:
		return fmt.Errorf("invalid phase: '%s'", p)
	}
	for _, hook := range hooks {
		if _, ok := executed[hook.Name()]; ok {
			glog.V(4).Infof("Skipping already executed hook for %s phase '%s'", p, hook.Name())
			continue
		}
		glog.V(4).Infof("Executing %s hook '%s'", p, hook.Name())
		if err := hook.Execute(pilot); err != nil {
			glog.V(4).Infof("Error executing %s hook '%s': %s", p, hook.Name(), err.Error())
			return fmt.Errorf("error executing %s hook '%s': %s", p, hook.Name(), err.Error())
		}
		glog.V(4).Infof("Executed %s hook '%s'", p, hook.Name())
		executed[hook.Name()] = struct{}{}
	}
	return nil
}
