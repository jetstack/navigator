package elastic

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	apps "k8s.io/client-go/informers/apps/v1beta1"
	depl "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
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
}

func NewElasticsearch(
	es informersv1.ElasticsearchClusterInformer,
	deploys depl.DeploymentInformer,
	statefulsets apps.StatefulSetInformer,
	cl *kubernetes.Clientset,
) *ElasticsearchController {
	elasticsearchController := &ElasticsearchController{
		kubeClient: cl,
	}

	es.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.addElasticsearchCluster,
		UpdateFunc: func(old, new interface{}) {
			logrus.Printf("Upd ES, old: %+v, new: %+v", old, new)
			// logic for dealing with a change to a TPR
		},
		DeleteFunc: func(obj interface{}) {
			logrus.Printf("Del ES")
			// logic for deleting an ES deployment
		},
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
func (e *ElasticsearchController) Run(stopCh <-chan struct{}) {
	logrus.Infof("Starting Elasticsearch controller")

	if !cache.WaitForCacheSync(stopCh, e.deployListerSynced, e.esListerSynced, e.statefulSetListerSynced) {
		util.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}

	<-stopCh
	logrus.Infof("Shutting down Elasticsearch controller")
}

func (e *ElasticsearchController) addElasticsearchCluster(obj interface{}) {
	var ok bool
	var es *v1.ElasticsearchCluster

	if es, ok = obj.(*v1.ElasticsearchCluster); !ok {
		logrus.Errorf("object not of type *v1.ElasticsearchCluster")
		return
	}

	e.needsUpdate(es)
}

func (e *ElasticsearchController) needsUpdate(c *v1.ElasticsearchCluster) (bool, error) {
	return false, nil
}
