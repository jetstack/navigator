package cassandra

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

// NewCassandra returns a new CassandraController that can be used
// to monitor for CassandraCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
type CassandraController struct {
	queue workqueue.RateLimitingInterface
}

// Run is the main event loop
func (e *CassandraController) Run(workers int, stopCh <-chan struct{}) error {
	glog.Infof("Starting Cassandra controller")

	// if !cache.WaitForCacheSync(stopCh, e.esListerSynced, e.pilotListerSynced, e.statefulSetListerSynced, e.podListerSynced, e.serviceAccountListerSynced, e.serviceListerSynced) {
	//	return fmt.Errorf("timed out waiting for caches to sync")
	// }

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

	if k, ok := key.(string); ok {
		glog.V(4).Infof("processing %q", k)
		e.queue.Forget(key)
	}

	return true
}

func NewCassandra() *CassandraController {
	queue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"cassandraCluster",
	)

	c := &CassandraController{
		queue: queue,
	}
	return c
}

func init() {
	controllers.Register("Cassandra", func(ctx *controllers.Context) controllers.Interface {
		e := NewCassandra()
		return e.Run
	})
}
