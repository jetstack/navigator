package elastic

import (
	"fmt"
	"reflect"

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
		AddFunc:    elasticsearchController.addElasticsearchCluster,
		UpdateFunc: elasticsearchController.updateElasticsearchCluster,
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

	if err := verifyElasticsearchCluster(es); err != nil {
		logrus.Errorf("error verifying ElasticsearchCluster resource: %s", err.Error())
		return
	}

	if needsUpdate, err := e.clusterNeedsUpdate(es); err != nil {
		logrus.Errorf("error checking if ElasticsearchCluster needs update: %s", err.Error())
		return
	} else if needsUpdate {
		logrus.Printf("Update ElasticsearchCluster!")
	}
}

func (e *ElasticsearchController) updateElasticsearchCluster(old, new interface{}) {
	// noop change
	if reflect.DeepEqual(old, new) {
		return
	}

	var ok bool
	var es *v1.ElasticsearchCluster

	if es, ok = new.(*v1.ElasticsearchCluster); !ok {
		logrus.Errorf("object not of type *v1.ElasticsearchCluster")
		return
	}

	if err := verifyElasticsearchCluster(es); err != nil {
		logrus.Errorf("error verifying ElasticsearchCluster resource: %s", err.Error())
		return
	}
}

func (e *ElasticsearchController) clusterNeedsUpdate(c *v1.ElasticsearchCluster) (bool, error) {
	for _, np := range c.Spec.NodePools {
		if needsUpdate, err := e.nodePoolNeedsUpdate(c, np); err != nil {
			return false, err
		} else if needsUpdate {
			return true, nil
		}
	}
	return false, nil
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
