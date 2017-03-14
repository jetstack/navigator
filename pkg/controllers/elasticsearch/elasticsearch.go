package elasticsearch

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	apps "k8s.io/client-go/informers/apps/v1beta1"
	depl "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
	"gitlab.jetstack.net/marshal/colonel/pkg/controllers"
	informersv1 "gitlab.jetstack.net/marshal/colonel/pkg/informers/v1"
	listersv1 "gitlab.jetstack.net/marshal/colonel/pkg/listers/v1"
	"gitlab.jetstack.net/marshal/colonel/pkg/util"
)

type ElasticsearchController struct {
	kubeClient *kubernetes.Clientset

	esLister       listersv1.ElasticsearchClusterLister
	esListerSynced cache.InformerSynced

	deployLister       extensionslisters.DeploymentLister
	deployListerSynced cache.InformerSynced

	statefulSetLister       appslisters.StatefulSetLister
	statefulSetListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewElasticsearch(
	es informersv1.ElasticsearchClusterInformer,
	deploys depl.DeploymentInformer,
	statefulsets apps.StatefulSetInformer,
	cl *kubernetes.Clientset,
) *ElasticsearchController {
	elasticsearchController := &ElasticsearchController{
		kubeClient: cl,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "elasticsearchCluster"),
	}

	es.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.enqueueElasticsearchCluster,
		UpdateFunc: func(old, cur interface{}) {
			oldCluster := old.(*v1.ElasticsearchCluster)
			curCluster := cur.(*v1.ElasticsearchCluster)
			if !reflect.DeepEqual(oldCluster, curCluster) {
				elasticsearchController.enqueueElasticsearchCluster(curCluster)
			}
		},
		DeleteFunc: elasticsearchController.enqueueElasticsearchCluster,
	})
	elasticsearchController.esLister = es.Lister()
	elasticsearchController.esListerSynced = es.Informer().HasSynced

	deploys.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logrus.Printf("Got deploy: %+v", obj)
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

	return elasticsearchController
}

// TODO: add a worker queue
func (e *ElasticsearchController) Run(workers int, stopCh <-chan struct{}) {
	defer e.queue.ShutDown()

	logrus.Infof("Starting Elasticsearch controller")

	if !cache.WaitForCacheSync(stopCh, e.deployListerSynced, e.esListerSynced, e.statefulSetListerSynced) {
		util.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}

	for i := 0; i < workers; i++ {
		go wait.Until(e.worker, time.Second, stopCh)
	}

	<-stopCh
	logrus.Infof("Shutting down Elasticsearch controller")
}

func (e *ElasticsearchController) worker() {
	for e.processNextWorkItem() {
	}
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

func (e *ElasticsearchController) sync(key string) error {
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
		logrus.Debugf("enqueued elasticsearchCluster '%s/%s' for update", es.Namespace, es.Name)
	}
	return nil
}

func (e *ElasticsearchController) enqueueElasticsearchCluster(obj interface{}) {
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
