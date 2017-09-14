package elasticsearch

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1beta1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	informerv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions/navigator/v1alpha1"
	listersv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/configmap"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/nodepool"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/service"
	"github.com/jetstack-experimental/navigator/pkg/controllers/elasticsearch/serviceaccount"
)

type ElasticsearchController struct {
	kubeClient kubernetes.Interface

	esLister       listersv1alpha1.ElasticsearchClusterLister
	esListerSynced cache.InformerSynced

	statefulSetLister       appslisters.StatefulSetLister
	statefulSetListerSynced cache.InformerSynced

	serviceAccountLister       corelisters.ServiceAccountLister
	serviceAccountListerSynced cache.InformerSynced

	serviceLister       corelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

	configMapLister       corelisters.ConfigMapLister
	configMapListerSynced cache.InformerSynced

	queue                       workqueue.RateLimitingInterface
	elasticsearchClusterControl ControlInterface
}

// NewElasticsearch returns a new ElasticsearchController that can be used
// to monitor for Elasticsearch resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
func NewElasticsearch(
	es cache.SharedIndexInformer,
	statefulsets cache.SharedIndexInformer,
	serviceaccounts cache.SharedIndexInformer,
	services cache.SharedIndexInformer,
	configmaps cache.SharedIndexInformer,
	cl kubernetes.Interface,
) *ElasticsearchController {
	// create an event broadcaster that can be used to send events to an event sink (eg. k8s)
	eventBroadcaster := record.NewBroadcaster()
	// log events to our logger
	eventBroadcaster.StartLogging(logrus.Infof)
	// log events to k8s
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: cl.Core().Events("")})
	recorder := eventBroadcaster.NewRecorder(api.Scheme, apiv1.EventSource{Component: "elasticsearchCluster"})

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "elasticsearchCluster")
	// create a new ElasticsearchController to manage ElasticsearchCluster resources
	elasticsearchController := &ElasticsearchController{
		kubeClient: cl,
		queue:      queue,
	}

	// add an event handler to the ElasticsearchCluster informer
	es.AddEventHandler(&controllers.QueuingEventHandler{Queue: queue})
	elasticsearchController.esLister = listersv1alpha1.NewElasticsearchClusterLister(es.GetIndexer())
	elasticsearchController.esListerSynced = es.HasSynced

	// add an event handler to the StatefulSet informer
	statefulsets.AddEventHandler(&controllers.BlockingEventHandler{WorkFunc: elasticsearchController.handleObject})
	elasticsearchController.statefulSetLister = appslisters.NewStatefulSetLister(statefulsets.GetIndexer())
	elasticsearchController.statefulSetListerSynced = statefulsets.HasSynced

	// add an event handler to the ServiceAccount informer
	serviceaccounts.AddEventHandler(&controllers.BlockingEventHandler{WorkFunc: elasticsearchController.handleObject})
	elasticsearchController.serviceAccountLister = corelisters.NewServiceAccountLister(serviceaccounts.GetIndexer())
	elasticsearchController.serviceAccountListerSynced = serviceaccounts.HasSynced

	// add an event handler to the Service informer
	services.AddEventHandler(&controllers.BlockingEventHandler{WorkFunc: elasticsearchController.handleObject})
	elasticsearchController.serviceLister = corelisters.NewServiceLister(services.GetIndexer())
	elasticsearchController.serviceListerSynced = services.HasSynced

	// add an event handler to the Service informer
	configmaps.AddEventHandler(&controllers.BlockingEventHandler{WorkFunc: elasticsearchController.handleObject})
	elasticsearchController.configMapLister = corelisters.NewConfigMapLister(configmaps.GetIndexer())
	elasticsearchController.configMapListerSynced = configmaps.HasSynced

	// create the actual ElasticsearchCluster controller
	elasticsearchController.elasticsearchClusterControl = NewController(
		elasticsearchController.statefulSetLister,
		elasticsearchController.serviceAccountLister,
		elasticsearchController.serviceLister,
		nodepool.NewController(
			cl,
			elasticsearchController.statefulSetLister,
			recorder,
		),
		configmap.NewController(
			cl,
			elasticsearchController.configMapLister,
			recorder,
		),
		serviceaccount.NewController(
			cl,
			elasticsearchController.serviceAccountLister,
			recorder,
		),
		service.NewController(
			cl,
			elasticsearchController.serviceLister,
			recorder,
		),
		recorder,
	)

	return elasticsearchController
}

// Run is the main event loop
func (e *ElasticsearchController) Run(workers int, stopCh <-chan struct{}) error {
	defer e.queue.ShutDown()

	logrus.Infof("Starting Elasticsearch controller")

	if !cache.WaitForCacheSync(stopCh, e.esListerSynced, e.statefulSetListerSynced, e.serviceAccountListerSynced, e.serviceListerSynced) {
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
	glog.V(4).Infof("Shutting down Elasticsearch controller workers...")
	wg.Wait()
	glog.V(4).Infof("Elasticsearch controller workers stopped.")
	return nil
}

func (e *ElasticsearchController) worker() {
	logrus.Infof("start worker loop")
	for e.processNextWorkItem() {
		logrus.Infof("processed work item")
	}
	logrus.Infof("exiting worker loop")
}

func (e *ElasticsearchController) processNextWorkItem() bool {
	key, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(key)

	if k, ok := key.(string); ok {
		if err := e.sync(k); err != nil {
			logrus.Infof("Error syncing ElasticsearchCluster %v, requeuing: %v", key.(string), err)
			e.queue.AddRateLimited(key)
		} else {
			e.queue.Forget(key)
		}
	}

	return true
}

func (e *ElasticsearchController) sync(key string) error {
	startTime := time.Now()
	defer func() {
		logrus.Infof("Finished syncing elasticsearchcluster %q (%v)", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	es, err := e.esLister.ElasticsearchClusters(namespace).Get(name)
	if errors.IsNotFound(err) {
		logrus.Infof("ElasticsearchCluster has been deleted %v", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve ElasticsearchCluster %v from store: %v", key, err))
		return err
	}

	return e.elasticsearchClusterControl.Sync(es)
}

func (e *ElasticsearchController) enqueueElasticsearchCluster(obj interface{}) {
	key, err := controllers.KeyFunc(obj)
	if err != nil {
		// TODO: log error
		logrus.Infof("Cound't get key for object %+v: %v", obj, err)
		return
	}
	logrus.Infof("Adding ES Cluster '%s' to queue", key)
	e.queue.Add(key)
}

func (e *ElasticsearchController) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		logrus.Errorf("error decoding object, invalid type")
		return
	}
	logrus.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		cluster, err := e.esLister.ElasticsearchClusters(object.GetNamespace()).Get(ownerRef.Name)

		if err != nil {
			logrus.Infof("ignoring orphaned object '%s' of elasticsearchcluster '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		e.enqueueElasticsearchCluster(cluster)
		return
	}
}

func init() {
	controllers.Register("ElasticSearch", func(ctx *controllers.Context) controllers.Interface {
		e := NewElasticsearch(
			ctx.SharedInformerFactory.InformerFor(
				ctx.Namespace,
				metav1.GroupVersionKind{Group: navigator.GroupName, Version: "v1alpha1", Kind: "ElasticsearchCluster"},
				informerv1alpha1.NewElasticsearchClusterInformer(ctx.NavigatorClient, ctx.Namespace, time.Second*30, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			),
			ctx.SharedInformerFactory.InformerFor(
				ctx.Namespace,
				metav1.GroupVersionKind{Group: "apps", Version: "v1beta1", Kind: "StatefulSet"},
				appsinformers.NewStatefulSetInformer(ctx.Client, ctx.Namespace, time.Second*30, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			),
			ctx.SharedInformerFactory.InformerFor(
				ctx.Namespace,
				metav1.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"},
				coreinformers.NewServiceAccountInformer(ctx.Client, ctx.Namespace, time.Second*30, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			),
			ctx.SharedInformerFactory.InformerFor(
				ctx.Namespace,
				metav1.GroupVersionKind{Version: "v1", Kind: "Service"},
				coreinformers.NewServiceInformer(ctx.Client, ctx.Namespace, time.Second*30, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			),
			ctx.SharedInformerFactory.InformerFor(
				ctx.Namespace,
				metav1.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
				coreinformers.NewConfigMapInformer(ctx.Client, ctx.Namespace, time.Second*30, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
			),
			ctx.Client,
		)

		return e.Run
	})
}
