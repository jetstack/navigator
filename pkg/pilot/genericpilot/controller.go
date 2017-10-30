package genericpilot

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/process"
)

const (
	ErrExecHook          = "ErrExecHook"
	ReasonProcessStarted = "ProcessStarted"

	MessageErrExecHook    = "Error executing hook: %s"
	MessageProcessStarted = "Subprocess (%s) started"
)

func (e *GenericPilot) worker() {
	glog.V(4).Infof("start worker loop")
	for e.processNextWorkItem() {
		glog.V(4).Infof("processed work item")
	}
	glog.V(4).Infof("exiting worker loop")
}

func (e *GenericPilot) processNextWorkItem() bool {
	key, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(key)

	if k, ok := key.(string); ok {
		if err := e.sync(k); err != nil {
			glog.Infof("Error syncing Pilot %v, requeuing: %v", key.(string), err)
			e.queue.AddRateLimited(key)
		} else {
			e.queue.Forget(key)
		}
	}

	return true
}

func (g *GenericPilot) sync(key string) (err error) {
	startTime := time.Now()
	defer func() {
		glog.Infof("Finished syncing pilot %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	pilot, err := g.pilotLister.Pilots(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.Infof("Pilot has been deleted %v", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve Pilot %v from store: %v", key, err))
		return err
	}

	// TODO: check namespace
	if g.Options.PilotName != pilot.Name {
		utilruntime.HandleError(fmt.Errorf("skipping pilot '%s'", pilot.Name))
		return nil
	}

	// TODO: ensure this pilot is of the correct type before syncing
	_, err = g.syncPilot(pilot)

	// TODO: enable status updates
	//updateErr := g.updateStatus(pilot, status)
	//if err == nil {
	//	return updateErr
	//}

	return err
}

func (g *GenericPilot) syncPilot(pilot *v1alpha1.Pilot) (v1alpha1.PilotStatus, error) {
	pilotCopy := pilot.DeepCopy()
	switch pilotCopy.Spec.Phase {
	case v1alpha1.PilotPhaseStarted:
		err := g.ensureProcessStarted(pilotCopy)
		if err != nil {
			// TODO: log Event & update status
			return pilotCopy.Status, err
		}
	}
	err := g.Options.SyncFunc(pilotCopy)
	return pilotCopy.Status, err
}

func (g *GenericPilot) updatePilotStatus(pilot *v1alpha1.Pilot) error {
	if g.process == nil || g.process.State() == nil || g.process.State().Exited() {
		pilot.UpdateStatusCondition(v1alpha1.PilotConditionStarted, v1alpha1.ConditionFalse, "", "")
	} else {
		pilot.UpdateStatusCondition(v1alpha1.PilotConditionStarted, v1alpha1.ConditionTrue, ReasonProcessStarted, MessageProcessStarted, g.process.String())
	}
	return nil
}

func (g *GenericPilot) ensureProcessStarted(pilot *v1alpha1.Pilot) error {
	if g.process == nil {
		err := g.constructProcess(pilot)
		if err != nil {
			return err
		}
	}

	if err := g.Options.Hooks.Transition(hook.PhasePreStart, pilot); err != nil {
		g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err.Error())
		return err
	}

	if state := g.process.State(); state != nil && state.Exited() == false {
		if err := g.Options.Hooks.Transition(hook.PhasePostStart, pilot); err != nil {
			g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err.Error())
			return err
		}
		return nil
	}

	if err := g.process.Start(); err != nil {
		glog.Fatalf("Error running child-process: %s", err.Error())
		return err
	}

	g.recorder.Eventf(pilot, corev1.EventTypeNormal, ReasonProcessStarted, MessageProcessStarted, g.process.String())

	go func() {
		if err := g.process.Wait(); err != nil {
			glog.Fatalf("Child-process exited with error: %s", err.Error())
		}
		glog.Fatalf("Child-process exited")
	}()

	return nil
}

func (g *GenericPilot) constructProcess(pilot *v1alpha1.Pilot) error {
	cmd, err := g.Options.CmdFunc(pilot)
	if err != nil {
		return err
	}
	g.process = &process.Adapter{
		Signals: g.Options.Signals,
		Cmd:     cmd,
	}
	return nil
}
