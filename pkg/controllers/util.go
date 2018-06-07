package controllers

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

// QueuingEventHandler is an implementation of cache.ResourceEventHandler that
// simply queues objects that are added/updated/deleted.
type QueuingEventHandler struct {
	Queue workqueue.RateLimitingInterface
}

func (q *QueuingEventHandler) Enqueue(obj interface{}) {
	key, err := KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	q.Queue.Add(key)
}

func (q *QueuingEventHandler) OnAdd(obj interface{}) {
	q.Enqueue(obj)
}

func (q *QueuingEventHandler) OnUpdate(old, new interface{}) {
	oldObj := old.(metav1.Object)
	newObj := new.(metav1.Object)
	if oldObj.GetResourceVersion() != newObj.GetResourceVersion() {
		q.Enqueue(new)
	}
}

func (q *QueuingEventHandler) OnDelete(obj interface{}) {
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = tombstone.Obj
	}
	q.Enqueue(obj)
}

// BlockingEventHandler is an implementation of cache.ResourceEventHandler that
// simply synchronously calls it's WorkFunc upon calls to OnAdd, OnUpdate or
// OnDelete.
type BlockingEventHandler struct {
	WorkFunc func(obj interface{})
}

func (b *BlockingEventHandler) Enqueue(obj interface{}) {
	b.WorkFunc(obj)
}

func (b *BlockingEventHandler) OnAdd(obj interface{}) {
	b.WorkFunc(obj)
}

func (b *BlockingEventHandler) OnUpdate(old, new interface{}) {
	oldObj := old.(metav1.Object)
	newObj := new.(metav1.Object)
	if oldObj.GetResourceVersion() != newObj.GetResourceVersion() {
		b.WorkFunc(new)
	}
}

func (b *BlockingEventHandler) OnDelete(obj interface{}) {
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = tombstone.Obj
	}
	b.WorkFunc(obj)
}

func RootControllerRef(state *State, objControlee runtime.Object) (*metav1.OwnerReference, error) {
	metaObj, err := meta.Accessor(objControlee)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get meta data for object")
	}
	ownerRef := metav1.GetControllerOf(metaObj)
	if ownerRef == nil {
		return ownerRef, nil
	}
	objName := metaObj.GetName()
	objNamespace := metaObj.GetNamespace()
	ownerName := ownerRef.Name

	var objController runtime.Object

	switch ownerRef.Kind {
	case "ElasticsearchCluster", "CassandraCluster":
		return ownerRef, nil
	case "StatefulSet":
		objController, err = state.StatefulSetLister.StatefulSets(objNamespace).Get(ownerName)
	case "Pod":
		objController, err = state.PodLister.Pods(objNamespace).Get(ownerName)
	default:
		return nil, fmt.Errorf(
			"object %s/%s has unsupported owner type: %v",
			objNamespace, objName, ownerRef,
		)
	}
	if err != nil {
		return nil, errors.Wrapf(
			err, "unable to get owner of object %s/%s/%s",
			objControlee.GetObjectKind().GroupVersionKind().Kind, objNamespace, objName,
		)
	}
	return RootControllerRef(state, objController)
}
