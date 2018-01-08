package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	ErrExecHook          = "ErrExecHook"
	ReasonProcessStarted = "ProcessStarted"
	ReasonPhaseComplete  = "ExecHook"

	MessageErrExecHook    = "Error executing hook: %s"
	MessageProcessStarted = "Subprocess %q started"
	MessagePhaseComplete  = "Completed phase %q"
)

func (e *Controller) worker() {
	glog.V(4).Infof("start worker loop")
	for e.processNextWorkItem() {
		glog.V(4).Infof("processed work item")
	}
	glog.V(4).Infof("exiting worker loop")
}

func (e *Controller) processNextWorkItem() bool {
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

func (g *Controller) sync(key string) (err error) {
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
		if !g.isThisPilot(name, namespace) {
			return nil
		}
		var thisPilot *v1alpha1.Pilot
		thisPilot, err = g.ThisPilot()
		if err != nil {
			return nil
		}
		glog.Infof("Using cached pilot resource for %q", key)
		pilot = thisPilot
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

	// store the most up to date copy of this pilot resource
	if g.IsThisPilot(pilot) {
		g.lock.Lock()
		g.cachedThisPilot = pilot
		g.lock.Unlock()
	}

	pilot = pilot.DeepCopy()
	// TODO: make 10 seconds configurable
	// we should resync all peers every 10s
	defer g.scheduledWorkQueue.Add(pilot, time.Second*10)
	err = g.syncFunc(pilot)
	if err != nil {
		return err
	}

	return nil
}
