package cassandra

import (
	"fmt"
	"sync"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	navigatorclientset "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	informerv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions/navigator/v1alpha1"
	listersv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/v1alpha1"

	"github.com/golang/glog"
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// NewCassandra returns a new CassandraController that can be used
// to monitor for CassandraCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
type CassandraController struct {
	navigatorClient         navigatorclientset.Interface
	cassandraClusterControl ControlInterface
	cassLister              listersv1alpha1.CassandraClusterLister
	cassListerSynced        cache.InformerSynced
	queue                   workqueue.RateLimitingInterface
}

func NewCassandra(
	navigatorClient navigatorclientset.Interface,
	ci cache.SharedIndexInformer,

) *CassandraController {
	queue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"cassandraCluster",
	)

	// add an event handler to the ElasticsearchCluster informer
	ci.AddEventHandler(&controllers.QueuingEventHandler{Queue: queue})

	return &CassandraController{
		navigatorClient:         navigatorClient,
		cassandraClusterControl: NewController(),
		cassLister: listersv1alpha1.NewCassandraClusterLister(
			ci.GetIndexer(),
		),
		cassListerSynced: ci.HasSynced,
		queue:            queue,
	}
}

// Run is the main event loop
func (e *CassandraController) Run(workers int, stopCh <-chan struct{}) error {
	glog.Infof("Starting Cassandra controller")

	if !cache.WaitForCacheSync(
		stopCh,
		e.cassListerSynced,
	) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.Until(e.worker, time.Second, stopCh)
		}()
	}

	<-stopCh
	e.queue.ShutDown()
	glog.V(4).Infof("Shutting down Cassandra controller workers...")
	wg.Wait()
	glog.V(4).Infof("Cassandra controller workers stopped.")
	return nil
}

func (e *CassandraController) worker() {
	glog.V(4).Infof("start worker loop")
	for e.processNextWorkItem() {
		glog.V(4).Infof("processed work item")
	}
	glog.V(4).Infof("exiting worker loop")
}

func (e *CassandraController) processNextWorkItem() bool {
	key, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(key)
	glog.V(4).Infof("processing %#v", key)
	if k, ok := key.(string); ok {
		if err := e.sync(k); err != nil {
			glog.Infof(
				"Error syncing CassandraCluster %v, requeuing: %v",
				key.(string), err,
			)
			e.queue.AddRateLimited(key)
		} else {
			e.queue.Forget(key)
		}
	} else {
		e.queue.Forget(key)
	}

	return true
}

func (e *CassandraController) sync(key string) (err error) {
	startTime := time.Now()
	defer func() {
		glog.Infof(
			"Finished syncing cassandracluster %q (%v)",
			key, time.Since(startTime),
		)
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	cass, err := e.cassLister.CassandraClusters(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.Infof("CassandraCluster has been deleted %v", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(
			fmt.Errorf(
				"unable to retrieve CassandraCluster %v from store: %v",
				key, err,
			),
		)
		return err
	}
	cass = cass.DeepCopy()
	status, err := e.cassandraClusterControl.Sync(cass)
	if err != nil {
		return err
	}
	glog.V(4).Infof("got status %#v", status)
	_, err = e.navigatorClient.
		NavigatorV1alpha1().
		CassandraClusters(cass.Namespace).
		UpdateStatus(cass)
	return err
}

func init() {
	controllers.Register("Cassandra", func(ctx *controllers.Context) controllers.Interface {
		e := NewCassandra(
			ctx.NavigatorClient,
			ctx.SharedInformerFactory.InformerFor(
				ctx.Namespace,
				metav1.GroupVersionKind{
					Group:   navigator.GroupName,
					Version: "v1alpha1",
					Kind:    "CassandraCluster",
				},
				informerv1alpha1.NewCassandraClusterInformer(
					ctx.NavigatorClient,
					ctx.Namespace,
					time.Second*30,
					cache.Indexers{
						cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
					},
				),
			),
		)
		return e.Run
	})
}
