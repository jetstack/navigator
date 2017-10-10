package genericpilot

import (
	"fmt"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/action"
)

func (g *GenericPilot) fireAction(actionName string, pilot *v1alpha1.Pilot) error {
	var a action.Interface
	var ok bool
	if a, ok = g.Options.Actions[actionName]; !ok {
		return fmt.Errorf("action '%s' not registered", actionName)
	}
	return a.Execute(pilot)
}
