package genericpilot

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/process"
)

const (
	ErrExecHook          = "ErrExecHook"
	ReasonProcessStarted = "ProcessStarted"
	ReasonPhaseComplete  = "ExecHook"

	MessageErrExecHook    = "Error executing hook: %s"
	MessageProcessStarted = "Subprocess %q started"
	MessagePhaseComplete  = "Completed phase %q"
)

func (e *GenericPilot) worker() {
	glog.V(4).Infof("start worker loop")
	for e.processNextWorkItem() {
		glog.V(4).Infof("processed work item")
	}
	glog.V(4).Infof("exiting worker loop")
}

func (e *GenericPilot) processNextWorkItem() bool {
	obj, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(obj)

	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		glog.Errorf("Unexpected non-string item in work queue: %#v", obj)
		e.queue.Forget(obj)
		return true
	}

	if err := e.sync(key); err != nil {
		glog.Infof("Error syncing Pilot %v, requeuing: %v", key, err)
		e.queue.AddRateLimited(key)
	} else {
		e.queue.Forget(obj)
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
	if apierrors.IsNotFound(err) {
		glog.Infof("Pilot %q has been deleted", key)
		if !g.isThisPilot(name, namespace) || g.cachedThisPilot == nil {
			return nil
		}
		glog.Infof("Using cached pilot resource for %q", key)
		pilot = g.cachedThisPilot
		// set err to nil so the following block does not return err
		err = nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve Pilot %v from store: %v", key, err))
		return err
	}

	// check if the pilot we are processing is a member of the same cluster as
	// this pilot
	isPeer, err := g.IsPeer(pilot)
	if err != nil {
		return err
	}
	if !isPeer {
		glog.V(2).Infof("Skipping pilot %q as it is not a peer in the cluster", pilot.Name)
		return nil
	}

	// TODO: make 10 seconds configurable
	// we should resync all peers every 10s
	defer g.scheduledWorkQueue.Add(pilot, time.Second*10)
	err = g.syncPilot(pilot)
	if err != nil {
		return err
	}

	return nil
}

func (g *GenericPilot) syncPilot(pilot *v1alpha1.Pilot) (err error) {
	// don't perform status updates for any Pilot other than our own
	if !g.IsThisPilot(pilot) {
		return g.Options.SyncFunc(pilot)
	}

	// store the most up to date copy of this pilot resource
	g.cachedThisPilot = pilot

	// we defer this to the end of execution to ensure it is run even if a part
	// of the sync errors
	defer func() {
		errs := []error{err}
		errs = append(errs, g.updatePilotStatus(pilot))
		// err will be nil if the contents of errs is all nil
		err = utilerrors.NewAggregate(errs)
	}()

	err = g.reconcileHooks(pilot)
	if err != nil {
		return
	}

	err = g.reconcileProcessState(pilot)
	if err != nil {
		return
	}

	err = g.Options.SyncFunc(pilot)
	if err != nil {
		return
	}

	return
}

func (g *GenericPilot) reconcileHooks(pilot *v1alpha1.Pilot) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	var phaseToExecute v1alpha1.PilotPhase
	switch g.lastCompletedPhase {
	case "":
		phaseToExecute = v1alpha1.PilotPhasePreStart
	case v1alpha1.PilotPhasePreStart:
		if !g.IsRunning() {
			glog.V(4).Infof("Not running post-start hooks as process has not started")
			return nil
		}
		phaseToExecute = v1alpha1.PilotPhasePostStart
	case v1alpha1.PilotPhasePostStart:
		if !g.shutdown {
			return nil
		}
		phaseToExecute = v1alpha1.PilotPhasePreStop
	case v1alpha1.PilotPhasePreStop:
		if g.IsRunning() {
			glog.V(4).Infof("Not running post-stop hooks as process is still running")
			return nil
		}
		phaseToExecute = v1alpha1.PilotPhasePostStop
	}
	if phaseToExecute == "" {
		glog.V(4).Infof("No phase hooks to execute")
		return nil
	}

	err := g.Options.Hooks.Transition(phaseToExecute, pilot)
	if err != nil {
		g.recorder.Eventf(pilot, corev1.EventTypeWarning, ErrExecHook, MessageErrExecHook, err.Error())
		return err
	}

	g.recorder.Eventf(pilot, corev1.EventTypeNormal, ReasonPhaseComplete, MessagePhaseComplete, phaseToExecute)
	g.lastCompletedPhase = phaseToExecute
	return nil
}

func (g *GenericPilot) reconcileProcessState(pilot *v1alpha1.Pilot) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.process == nil {
		err := g.constructProcess(pilot)
		if err != nil {
			return err
		}
	}
	switch {
	case !g.IsRunning() && g.lastCompletedPhase == v1alpha1.PilotPhasePreStart:
		err := g.process.Start()
		if err != nil {
			return err
		}
		g.recorder.Eventf(pilot, corev1.EventTypeNormal, ReasonProcessStarted, MessageProcessStarted, g.process.String())
	case g.IsRunning() && g.lastCompletedPhase == v1alpha1.PilotPhasePreStop:
		err := g.process.Stop()
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

	pilot.Status.LastCompletedPhase = g.lastCompletedPhase

	// perform update in API
	_, err := g.client.NavigatorV1alpha1().Pilots(pilot.Namespace).UpdateStatus(pilot)
	return err
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
