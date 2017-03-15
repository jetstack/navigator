package elasticsearch

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	apps "k8s.io/client-go/informers/apps/v1beta1"
	depl "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/api"
	clientv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
	"gitlab.jetstack.net/marshal/colonel/pkg/controllers"
	informersv1 "gitlab.jetstack.net/marshal/colonel/pkg/informers/v1"
	listersv1 "gitlab.jetstack.net/marshal/colonel/pkg/listers/v1"
)

type ElasticsearchController struct {
	kubeClient *kubernetes.Clientset

	esLister       listersv1.ElasticsearchClusterLister
	esListerSynced cache.InformerSynced

	deployLister       extensionslisters.DeploymentLister
	deployListerSynced cache.InformerSynced

	statefulSetLister       appslisters.StatefulSetLister
	statefulSetListerSynced cache.InformerSynced

	queue                       workqueue.RateLimitingInterface
	elasticsearchClusterControl ElasticsearchClusterControl
}

func NewElasticsearch(
	es informersv1.ElasticsearchClusterInformer,
	deploys depl.DeploymentInformer,
	statefulsets apps.StatefulSetInformer,
	cl *kubernetes.Clientset,
) *ElasticsearchController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(cl.Core().RESTClient()).Events("")})
	recorder := eventBroadcaster.NewRecorder(api.Scheme, clientv1.EventSource{Component: "elasticsearchCluster"})

	elasticsearchController := &ElasticsearchController{
		kubeClient: cl,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "elasticsearchCluster"),
	}

	es.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.enqueueElasticsearchCluster,
		UpdateFunc: func(old, cur interface{}) {
			logrus.Printf("edited esc")
			oldCluster := old.(*v1.ElasticsearchCluster)
			curCluster := cur.(*v1.ElasticsearchCluster)
			if !reflect.DeepEqual(oldCluster, curCluster) {
				logrus.Printf("queue update")
				elasticsearchController.enqueueElasticsearchCluster(curCluster)
			}
		},
		DeleteFunc: elasticsearchController.enqueueElasticsearchCluster,
	})
	elasticsearchController.esLister = es.Lister()
	elasticsearchController.esListerSynced = es.Informer().HasSynced

	deploys.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// can be mostly ignored
		},
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			logrus.Printf("edited deploy")
			// did our deployment change? if so, does it break a contract?
			// maybe we should reconcile those problems here. we should be careful to
			// avoid infinite loops in case of bugs with colonel however
		},
		DeleteFunc: func(obj interface{}) {
			// if our deployment has been deleted, we should re-create it
		},
	})
	elasticsearchController.deployLister = deploys.Lister()
	elasticsearchController.deployListerSynced = deploys.Informer().HasSynced

	statefulsets.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// can be mostly ignored
		},
		UpdateFunc: func(old, new interface{}) {
			// did our deployment change? if so, does it break a contract?
			// maybe we should reconcile those problems here. we should be careful to
			// avoid infinite loops in case of bugs with colonel however
		},
		DeleteFunc: func(obj interface{}) {
			// if our deployment has been deleted, we should re-create it
		},
	})
	elasticsearchController.statefulSetLister = statefulsets.Lister()
	elasticsearchController.statefulSetListerSynced = statefulsets.Informer().HasSynced

	elasticsearchController.elasticsearchClusterControl = NewElasticsearchClusterControl(
		NewElasticsearchClusterNodePoolControl(
			cl,
			elasticsearchController.statefulSetLister,
			elasticsearchController.deployLister,
			recorder,
		),
	)

	return elasticsearchController
}

func (e *ElasticsearchController) Run(workers int, stopCh <-chan struct{}) {
	defer e.queue.ShutDown()

	logrus.Infof("Starting Elasticsearch controller")

	if !cache.WaitForCacheSync(stopCh, e.deployListerSynced, e.esListerSynced, e.statefulSetListerSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}

	for i := 0; i < workers; i++ {
		go wait.Until(e.worker, time.Second, stopCh)
	}

	<-stopCh
	logrus.Infof("Shutting down Elasticsearch controller")
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
	if err := e.sync(key.(string)); err != nil {
		logrus.Infof("Error syncing ElasticsearchCluster %v, requeuing: %v", key.(string), err)
		e.queue.AddRateLimited(key)
	} else {
		e.queue.Forget(key)
	}
	return true
}

// TODO: properly log errors to an events sink
// TODO: move verification out of this function
func (e *ElasticsearchController) sync(key string) error {
	logrus.Debugf("syncing '%s'", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	es, err := e.esLister.ElasticsearchClusters(namespace).Get(name)
	if errors.IsNotFound(err) {
		logrus.Infof("ElasticsearchCluster has been deleted: %v", key)
		return nil
	}

	if err != nil {
		logrus.Errorf("unable to retreive ElasticsearchCluster from store: %v", err.Error())
		return err
	}

	if err := verifyElasticsearchCluster(es); err != nil {
		logrus.Errorf("error verifying ElasticsearchCluster resource: %s", err.Error())
		return err
	}

	if needsUpdate, err := e.clusterNeedsUpdate(es); err != nil {
		logrus.Errorf("error checking if ElasticsearchCluster needs update: %s", err.Error())
		return err
	} else if needsUpdate {
		logrus.Debugf("syncing ElasticsearchCluster '%s/%s'", es.Namespace, es.Name)
		if err := e.elasticsearchClusterControl.SyncElasticsearchCluster(es); err != nil {
			logrus.Errorf("error syncing ElasticsearchCluster: %s", err.Error())
			return err
		}
	}

	return nil
}

func (e *ElasticsearchController) enqueueElasticsearchCluster(obj interface{}) {
	logrus.Debugf("enqueuing object: %+v", obj)
	key, err := controllers.KeyFunc(obj)
	if err != nil {
		// TODO: log error
		logrus.Infof("Cound't get key for object %+v: %v", obj, err)
		return
	}
	e.queue.Add(key)
}

func verifyElasticsearchCluster(c *v1.ElasticsearchCluster) error {
	// TODO: add verification that at least one client, master and data node pool exist
	if c.Spec.Version == "" {
		return fmt.Errorf("cluster version number must be specified")
	}

	for _, np := range c.Spec.NodePools {
		if err := verifyNodePool(np); err != nil {
			return err
		}
	}

	return nil
}

func verifyNodePool(np *v1.ElasticsearchClusterNodePool) error {
	for _, role := range np.Roles {
		switch role {
		case "data", "client", "master":
		default:
			return fmt.Errorf("invalid role '%s' specified. must be one of 'data', 'client' or 'master'", role)
		}
	}

	if np.State != nil {
		if !np.State.Stateful && np.State.Persistence.Enabled {
			return fmt.Errorf("a non-stateful node pool cannot have persistence enabled")
		}
	}

	return nil
}
