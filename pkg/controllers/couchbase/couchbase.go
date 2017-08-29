package couchbase

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps"
	"k8s.io/client-go/pkg/apis/extensions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	appsinformers "k8s.io/client-go/informers/apps/v1beta1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	depl "k8s.io/client-go/informers/extensions/v1beta1"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"

	informerv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions/navigator/v1alpha1"
	listersv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/v1alpha1"
)

type CouchbaseController struct {
	kubeClient *kubernetes.Clientset

	cbLister       listersv1alpha1.CouchbaseClusterLister
	cbListerSynced cache.InformerSynced

	deployLister       extensionslisters.DeploymentLister
	deployListerSynced cache.InformerSynced

	statefulSetLister       appslisters.StatefulSetLister
	statefulSetListerSynced cache.InformerSynced

	serviceAccountLister       corelisters.ServiceAccountLister
	serviceAccountListerSynced cache.InformerSynced

	serviceLister       corelisters.ServiceLister
	serviceListerSynced cache.InformerSynced

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

func (c *CouchbaseController) handleDeploy(obj interface{}) {
	var deploy *extensions.Deployment
	var ok bool
	if deploy, ok = obj.(*extensions.Deployment); !ok {
		logrus.Errorf("error decoding deployment, invalid type")
		return
	}
	if ownerRef := managedOwnerRef(deploy.ObjectMeta); ownerRef != nil {
		logrus.Debugf("getting couchbasecluster '%s/%s'", deploy.Namespace, ownerRef.Name)
		cluster, err := c.cbLister.CouchbaseClusters(deploy.Namespace).Get(ownerRef.Name)

		if err != nil {
			logrus.Infof("ignoring orphaned deployment '%s' of couchbasecluster '%s'", deploy.Name, ownerRef.Name)
			return
		}

		c.enqueueCouchbaseCluster(cluster)
		return
	}
}

func (c *CouchbaseController) handleStatefulSet(obj interface{}) {
	var ss *apps.StatefulSet
	var ok bool
	if ss, ok = obj.(*apps.StatefulSet); !ok {
		logrus.Errorf("error decoding statefulset, invalid type")
		return
	}
	if ownerRef := managedOwnerRef(ss.ObjectMeta); ownerRef != nil {
		cluster, err := c.cbLister.CouchbaseClusters(ss.Namespace).Get(ownerRef.Name)

		if err != nil {
			logrus.Infof("ignoring orphaned statefulset '%s' of couchbasecluster '%s'", ss.Name, ownerRef.Name)
			return
		}

		c.enqueueCouchbaseCluster(cluster)
		return
	}
}

func (c *CouchbaseController) handleServiceAccount(obj interface{}) {
	var ss *apiv1.ServiceAccount
	var ok bool
	if ss, ok = obj.(*apiv1.ServiceAccount); !ok {
		logrus.Errorf("error decoding serviceaccount, invalid type")
		return
	}
	if ownerRef := managedOwnerRef(ss.ObjectMeta); ownerRef != nil {
		cluster, err := c.cbLister.CouchbaseClusters(ss.Namespace).Get(ownerRef.Name)

		if err != nil {
			logrus.Infof("ignoring orphaned serviceaccount '%s' of couchbasecluster '%s'", ss.Name, ownerRef.Name)
			return
		}

		c.enqueueCouchbaseCluster(cluster)
		return
	}
}

func (c *CouchbaseController) handleService(obj interface{}) {
	var ss *apiv1.Service
	var ok bool
	if ss, ok = obj.(*apiv1.Service); !ok {
		logrus.Errorf("error decoding service, invalid type")
		return
	}
	if ownerRef := managedOwnerRef(ss.ObjectMeta); ownerRef != nil {
		cluster, err := c.cbLister.CouchbaseClusters(ss.Namespace).Get(ownerRef.Name)

		if err != nil {
			logrus.Infof("ignoring orphaned service '%s' of couchbasecluster '%s'", ss.Name, ownerRef.Name)
			return
		}

		c.enqueueCouchbaseCluster(cluster)
		return
	}
}

// NewCouchbase returns a new CouchbaseController that can be used
// to monitor for CouchbaseCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
func NewCouchbase(
	cbInformer informerv1alpha1.CouchbaseClusterInformer,
	deploys depl.DeploymentInformer,
	statefulsets appsinformers.StatefulSetInformer,
	serviceaccounts coreinformers.ServiceAccountInformer,
	services coreinformers.ServiceInformer,
	cl *kubernetes.Clientset,
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

	// create a new ElasticsearchController to manage ElasticsearchCluster resources
	cbController := &CouchbaseController{
		kubeClient: cl,
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"couchbaseCluster",
		),
	}

	// add an event handler to the ElasticsearchCluster informer
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

	// add an event handler to the Deployment informer
	deploys.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cbController.handleDeploy,
		UpdateFunc: func(old, cur interface{}) {
			if reflect.DeepEqual(old, cur) {
				return
			}
			cbController.handleDeploy(cur)
		},
		DeleteFunc: func(obj interface{}) {
			cbController.handleDeploy(obj)
		},
	})
	cbController.deployLister = deploys.Lister()
	cbController.deployListerSynced = deploys.Informer().HasSynced

	// add an event handler to the StatefulSet informer
	statefulsets.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cbController.handleStatefulSet,
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			cbController.handleStatefulSet(new)
		},
		DeleteFunc: cbController.handleStatefulSet,
	})
	cbController.statefulSetLister = statefulsets.Lister()
	cbController.statefulSetListerSynced = statefulsets.Informer().HasSynced

	// add an event handler to the ServiceAccount informer
	serviceaccounts.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cbController.handleServiceAccount,
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			cbController.handleServiceAccount(new)
		},
		DeleteFunc: cbController.handleServiceAccount,
	})
	cbController.serviceAccountLister = serviceaccounts.Lister()
	cbController.serviceAccountListerSynced = serviceaccounts.Informer().HasSynced

	// add an event handler to the Service informer
	services.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cbController.handleService,
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			cbController.handleService(new)
		},
		DeleteFunc: cbController.handleService,
	})
	cbController.serviceLister = services.Lister()
	cbController.serviceListerSynced = services.Informer().HasSynced

	// create the actual CouchbaseCluster controller
	cbController.cbClusterControl = NewCouchbaseClusterControl(
		cbController.statefulSetLister,
		cbController.deployLister,
		cbController.serviceAccountLister,
		cbController.serviceLister,
		NewCouchbaseClusterNodePoolControl(
			cl,
			cbController.deployLister,
			recorder,
		),
		NewStatefulCouchbaseClusterNodePoolControl(
			cl,
			cbController.statefulSetLister,
			recorder,
		),
		NewCouchbaseClusterServiceAccountControl(
			cl,
			recorder,
		),
		// client service controller
		NewCouchbaseClusterServiceControl(
			cl,
			recorder,
			ServiceControlConfig{
				NameSuffix: "clients",
				EnableHTTP: true,
				Roles:      []string{"client"},
			},
		),
		// discovery service controller
		NewCouchbaseClusterServiceControl(
			cl,
			recorder,
			ServiceControlConfig{
				NameSuffix:  "discovery",
				Annotations: map[string]string{"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true"},
			},
		),
		recorder,
	)

	return cbController
}

func init() {
	controllers.Register("Couchbase", func(ctx *controllers.Context) (bool, error) {
		go NewCouchbase(
			ctx.NavigatorInformerFactory.Navigator().V1alpha1().CouchbaseClusters(),
			ctx.InformerFactory.Extensions().V1beta1().Deployments(),
			ctx.InformerFactory.Apps().V1beta1().StatefulSets(),
			ctx.InformerFactory.Core().V1().ServiceAccounts(),
			ctx.InformerFactory.Core().V1().Services(),
			ctx.Client,
		).Run(2, ctx.Stop)

		return true, nil
	})
}
