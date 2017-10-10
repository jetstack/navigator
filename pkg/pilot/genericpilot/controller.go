package genericpilot

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/action"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/process"
)

const (
	ErrExecHook = "ErrExecHook"

	MessageErrExecHook = "Error executing hook: %s"
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
			glog.Infof("Error syncing ElasticsearchCluster %v, requeuing: %v", key.(string), err)
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
		glog.Infof("Finished syncing elasticsearchcluster %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	pilot, err := g.pilotLister.Pilots(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.Infof("ElasticsearchCluster has been deleted %v", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve ElasticsearchCluster %v from store: %v", key, err))
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
	if pilot.Spec.Phase == v1alpha1.PilotPhaseStarted {
		err := g.ensureProcessStarted(pilotCopy)
		if err != nil {
			// TODO: log Event & update status
			return pilotCopy.Status, err
		}
	}
	if pilot.Spec.Phase == v1alpha1.PilotPhaseDecommissioned {
		err := g.fireAction(action.Decommission, pilotCopy)
		if err != nil {
			// TODO: log event & update status
			return pilotCopy.Status, err
		}
	}
	err := g.Options.SyncFunc(pilotCopy)
	return pilotCopy.Status, err
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

	// TODO: block here until the process has started and update the pilot.Status block accordingly
	// TODO: trigger a resync here/notify the pilot consumer that the sub-process has exited.
	go func() {
		err := g.process.Run()
		if err != nil {
			glog.Fatalf("Error running child-process: %s", err.Error())
		}
		glog.Fatalf("Process exited")
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
		Stdout:  g.Options.Stdout,
		Stderr:  g.Options.Stderr,
		Cmd:     cmd,
	}
	return nil
}
