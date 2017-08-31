package couchbase

import (
	"fmt"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	informerv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions/navigator/v1alpha1"
	listersv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/v1alpha1"
)

type CouchbaseController struct {
	kubeClient kubernetes.Interface

	cbLister       listersv1alpha1.CouchbaseClusterLister
	cbListerSynced cache.InformerSynced

	queue                   workqueue.RateLimitingInterface
	couchbaseClusterControl CouchbaseClusterControl
}

func (c *CouchbaseController) enqueueCouchbaseCluster(obj interface{}) {
	key, err := controllers.KeyFunc(obj)
	if err != nil {
		// TODO: log error
		logrus.Infof("Cound't get key for object %+v: %v", obj, err)
		return
	}
	c.queue.Add(key)
}

func (c *CouchbaseController) enqueueCouchbaseClusterDelete(obj interface{}) {
	c.queue.Add(obj)
}

// NewCouchbase returns a new CouchbaseController that can be used
// to monitor for CouchbaseCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
func NewCouchbase(
	cbInformer informerv1alpha1.CouchbaseClusterInformer,
	cl kubernetes.Interface,
) *CouchbaseController {
	// create an event broadcaster that can be used to send events to an event sink (eg. k8s)
	eventBroadcaster := record.NewBroadcaster()
	// log events to our logger
	eventBroadcaster.StartLogging(logrus.Infof)
	// log events to k8s
	eventBroadcaster.StartRecordingToSink(
		&v1core.EventSinkImpl{
			Interface: cl.Core().Events(""),
		},
	)
	recorder := eventBroadcaster.NewRecorder(
		api.Scheme,
		apiv1.EventSource{
			Component: "couchbaseCluster",
		},
	)

	cbController := &CouchbaseController{
		kubeClient: cl,
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"couchbaseCluster",
		),
	}

	cbInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cbController.enqueueCouchbaseCluster,
		UpdateFunc: func(old, cur interface{}) {
			if reflect.DeepEqual(old, cur) {
				return
			}
			cbController.enqueueCouchbaseCluster(cur)
		},
		DeleteFunc: cbController.enqueueCouchbaseClusterDelete,
	})
	cbController.cbLister = cbInformer.Lister()
	cbController.cbListerSynced = cbInformer.Informer().HasSynced

	cbController.couchbaseClusterControl = NewCouchbaseClusterControl(recorder)

	return cbController
}

// Run is the main event loop
func (c *CouchbaseController) Run(workers int, stopCh <-chan struct{}) {
	defer c.queue.ShutDown()

	logrus.Infof("Starting Couchbase controller")

	if !cache.WaitForCacheSync(stopCh, c.cbListerSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	logrus.Infof("Shutting down Couchbase controller")
}

func (c *CouchbaseController) worker() {
	logrus.Infof("start worker loop")
	for c.processNextWorkItem() {
		logrus.Infof("processed work item")
	}
	logrus.Infof("exiting worker loop")
}

func (c *CouchbaseController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	if k, ok := key.(string); ok {
		if err := c.sync(k); err != nil {
			logrus.Infof("Error syncing CouchbaseCluster %v, requeuing: %v", key.(string), err)
			c.queue.AddRateLimited(key)
		} else {
			c.queue.Forget(key)
		}
	} else if cb, ok := key.(*v1alpha1.CouchbaseCluster); ok {
		t := metav1.NewTime(time.Now())
		cb.DeletionTimestamp = &t
		if err := c.couchbaseClusterControl.SyncCouchbaseCluster(*cb); err != nil {
			logrus.Infof("Error syncing CouchbaseCluster %v, requeuing: %v", cb.Name, err)
		}
		c.queue.Forget(key)
	}

	return true
}

func (c *CouchbaseController) sync(key string) error {
	startTime := time.Now()
	defer func() {
		logrus.Infof("Finished syncing couchbasecluster %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	es, err := c.cbLister.CouchbaseClusters(namespace).Get(name)
	if errors.IsNotFound(err) {
		logrus.Infof("CouchbaseCluster has been deleted %v", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve CouchbaseCluster %v from store: %v", key, err))
		return err
	}

	return c.couchbaseClusterControl.SyncCouchbaseCluster(*es)
}

func init() {
	controllers.Register("Couchbase", func(ctx *controllers.Context) (bool, error) {
		go NewCouchbase(
			ctx.NavigatorInformerFactory.Navigator().V1alpha1().CouchbaseClusters(),
			ctx.Client,
		).Run(2, ctx.Stop)

		return true, nil
	})
}
