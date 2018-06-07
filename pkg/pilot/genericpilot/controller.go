package genericpilot

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/processmanager"
)

const (
	ErrExecHook          = "ErrExecHook"
	ReasonProcessStarted = "ProcessStarted"
	ReasonPhaseComplete  = "ExecHook"

	MessageErrExecHook    = "Error executing hook: %s"
	MessageProcessStarted = "Subprocess %q started"
	MessagePhaseComplete  = "Completed phase %q"
)

func (g *GenericPilot) syncPilot(pilot *v1alpha1.Pilot) (err error) {
	// we defer this to the end of execution to ensure it is run even if a part
	// of the sync errors
	defer func() {
		if g.IsThisPilot(pilot) {
			errs := []error{err}
			errs = append(errs, g.updatePilotStatus(pilot))
			// err will be nil if the contents of errs is all nil
			err = utilerrors.NewAggregate(errs)
		}
	}()

	g.lock.Lock()
	defer g.lock.Unlock()

	// we defer this to ensure it is run regardless of whether the pilot has
	// actually started, else progress across the rest of the application may
	// be stalled
	defer func() {
		if g.Options.LeaderElectedSyncFunc != nil && g.elector.Leading() {
			errs := []error{err}
			errs = append(errs, g.Options.LeaderElectedSyncFunc(pilot))
			err = utilerrors.NewAggregate(errs)
		}
	}()

	if !g.IsThisPilot(pilot) {
		return
	}

	err = g.Options.Hooks.Transition(hook.PreStart, pilot)
	if err != nil {
		g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err)
		return err
	}

	if g.process == nil {
		err := g.constructProcess(pilot)
		if err != nil {
			return err
		}
	}

	if !g.IsRunning() && !g.shutdown {
		err := g.process.Start()
		if err != nil {
			return err
		}
		g.recorder.Eventf(pilot, corev1.EventTypeNormal, ReasonProcessStarted, MessageProcessStarted, g.process.String())
	}

	// TODO: do we need to check if process is running here, or wait at all? What if the process is running and then exits quickly?
	err = g.Options.Hooks.Transition(hook.PostStart, pilot)
	if err != nil {
		g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err)
		return err
	}

	err = g.Options.SyncFunc(pilot)
	return
}

func (g *GenericPilot) stop(pilot *v1alpha1.Pilot) error {
	g.shutdown = true
	glog.V(4).Infof("Waiting for process exit and hooks to execute")

	if !g.IsRunning() {
		return nil
	}

	err := g.Options.Hooks.Transition(hook.PreStop, pilot)
	if err != nil {
		g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err)
		return err
	}

	// TODO: fire postStop hooks even if the subprocess failed
	err = g.process.Stop()
	if err != nil {
		return err
	}

	err = g.Options.Hooks.Transition(hook.PostStop, pilot)
	if err != nil {
		g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err)
		return err
	}

	return nil
}

// updatePilotStatus will update a pilots status field. It *will* mutate the
// Pilot it is passed as an argument, so ensure you have performed a DeepCopy
// before passing a pilot to this function.
func (g *GenericPilot) updatePilotStatus(pilot *v1alpha1.Pilot) error {
	// Set process started status condition
	if !g.IsRunning() {
		pilot.UpdateStatusCondition(v1alpha1.PilotConditionStarted, v1alpha1.ConditionFalse, "", "")
	} else {
		pilot.UpdateStatusCondition(v1alpha1.PilotConditionStarted, v1alpha1.ConditionTrue, ReasonProcessStarted, MessageProcessStarted, g.process.String())
	}

	// perform update in API
	_, err := g.client.NavigatorV1alpha1().Pilots(pilot.Namespace).UpdateStatus(pilot)
	return errors.Wrap(err, "unable to update pilot status")
}

func (g *GenericPilot) constructProcess(pilot *v1alpha1.Pilot) error {
	cmd, err := g.Options.CmdFunc(pilot)
	if err != nil {
		return err
	}
	g.process = processmanager.New(cmd, g.Options.Signals)
	return nil
}
