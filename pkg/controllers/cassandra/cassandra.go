package cassandra

import (
	"fmt"
	"sync"
	"time"

	navigatorinformers "github.com/jetstack/navigator/pkg/client/informers/externalversions/navigator/v1alpha1"

	"github.com/golang/glog"
	navigatorclientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	servicecql "github.com/jetstack/navigator/pkg/controllers/cassandra/service/cql"
	serviceseedprovider "github.com/jetstack/navigator/pkg/controllers/cassandra/service/seedprovider"
	appsinformers "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions/apps/v1beta1"
	coreinformers "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

// NewCassandra returns a new CassandraController that can be used
// to monitor for CassandraCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
type CassandraController struct {
	control                 ControlInterface
	cassLister              listersv1alpha1.CassandraClusterLister
	cassListerSynced        cache.InformerSynced
	serviceListerSynced     cache.InformerSynced
	statefulSetListerSynced cache.InformerSynced
	pilotsListerSynced      cache.InformerSynced
	podsListerSynced        cache.InformerSynced
	queue                   workqueue.RateLimitingInterface
	recorder                record.EventRecorder
}

func NewCassandra(
	naviClient navigatorclientset.Interface,
	kubeClient kubernetes.Interface,
	cassClusters navigatorinformers.CassandraClusterInformer,
	services coreinformers.ServiceInformer,
	statefulSets appsinformers.StatefulSetInformer,
	pilots navigatorinformers.PilotInformer,
	pods coreinformers.PodInformer,
	recorder record.EventRecorder,
) *CassandraController {
	queue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"cassandraCluster",
	)

	cc := &CassandraController{
		queue:    queue,
		recorder: recorder,
	}
	cassClusters.Informer().AddEventHandler(
		&controllers.QueuingEventHandler{Queue: queue},
	)
	cc.cassLister = cassClusters.Lister()
	cc.cassListerSynced = cassClusters.Informer().HasSynced
	cc.serviceListerSynced = services.Informer().HasSynced
	cc.statefulSetListerSynced = statefulSets.Informer().HasSynced
	cc.pilotsListerSynced = pilots.Informer().HasSynced
	cc.podsListerSynced = pods.Informer().HasSynced
	cc.control = NewControl(
		serviceseedprovider.NewControl(
			kubeClient,
			services.Lister(),
			recorder,
		),
		servicecql.NewControl(
			kubeClient,
			services.Lister(),
			recorder,
		),
		nodepool.NewControl(
			kubeClient,
			statefulSets.Lister(),
			recorder,
		),
		pilot.NewControl(
			naviClient,
			pilots.Lister(),
			pods.Lister(),
			statefulSets.Lister(),
			recorder,
		),
		recorder,
	)
	cc.recorder = recorder
	return cc
}

// Run is the main event loop
func (e *CassandraController) Run(workers int, stopCh <-chan struct{}) error {
	glog.Infof("Starting Cassandra controller")

	if !cache.WaitForCacheSync(
		stopCh,
		e.cassListerSynced,
		e.serviceListerSynced,
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

	if err := e.sync(key.(string)); err != nil {
		glog.Infof(
			"Error syncing CassandraCluster %v, requeuing: %v",
			key.(string), err,
		)
		e.queue.AddRateLimited(key)
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
	return e.control.Sync(cass.DeepCopy())
}

func (e *CassandraController) handleObject(obj interface{}) {
}

func CassandraControllerFromContext(ctx *controllers.Context) *CassandraController {
	return NewCassandra(
		ctx.NavigatorClient,
		ctx.Client,
		ctx.SharedInformerFactory.Navigator().V1alpha1().CassandraClusters(),
		ctx.KubeSharedInformerFactory.Core().V1().Services(),
		ctx.KubeSharedInformerFactory.Apps().V1beta1().StatefulSets(),
		ctx.SharedInformerFactory.Navigator().V1alpha1().Pilots(),
		ctx.KubeSharedInformerFactory.Core().V1().Pods(),
		ctx.Recorder,
	)
}

func init() {
	controllers.Register(
		"Cassandra",
		func(ctx *controllers.Context) controllers.Interface {
			return CassandraControllerFromContext(ctx).Run
		},
	)
}
