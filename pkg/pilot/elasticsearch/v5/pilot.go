package v5

import (
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/action"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/periodic"
)

type Pilot struct {
	Options Options
}

func (p *Pilot) ConfigureGenericPilot(opts *genericpilot.Options) {
	opts.CmdFunc = p.CmdFunc
	opts.SyncFunc = p.syncFunc
	opts.Actions = p.Actions()
	opts.Periodics = p.Periodics()
	opts.Hooks = p.Hooks()
}

func (p *Pilot) Hooks() *hook.Hooks {
	return &hook.Hooks{
		PreStart: []hook.Interface{
			hook.New("InstallPlugins", p.InstallPlugins),
		},
	}
}

func (p *Pilot) Actions() map[string]action.Interface {
	return map[string]action.Interface{
		action.Decommission: action.New(action.Decommission, p.actionDecommission),
	}
}

func (p *Pilot) Periodics() map[string]periodic.Interface {
	return map[string]periodic.Interface{
		periodic.Decommission: periodic.New(periodic.Decommission, p.periodicDecommission),
	}
}
