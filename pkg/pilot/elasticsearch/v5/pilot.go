package v5

import (
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/action"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/periodic"
)

type Pilot struct {
	Options Options
}

func ConfigureGenericPilot(opts *Options) {
	p := &Pilot{Options: *opts}
	opts.GenericPilotOptions.CmdFunc = p.CmdFunc
	opts.GenericPilotOptions.SyncFunc = p.syncFunc
	opts.GenericPilotOptions.Periodics = p.Periodics()
	opts.GenericPilotOptions.Hooks = p.Hooks()
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
