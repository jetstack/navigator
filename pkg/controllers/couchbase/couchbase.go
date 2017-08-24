package couchbase

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/navigator/pkg/controllers"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	apiv1 "k8s.io/client-go/pkg/api/v1"
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

// NewCouchbase returns a new CouchbaseController that can be used
// to monitor for CouchbaseCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
func NewCouchbase(
	es informerv1alpha1.CouchbaseClusterInformer,
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
			Component: "elasticsearchCluster",
		},
	)

	// create a new ElasticsearchController to manage ElasticsearchCluster resources
	elasticsearchController := &CouchbaseController{
		kubeClient: cl,
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"elasticsearchCluster",
		),
	}

	// add an event handler to the ElasticsearchCluster informer
	es.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.enqueueElasticsearchCluster,
		UpdateFunc: func(old, cur interface{}) {
			if reflect.DeepEqual(old, cur) {
				return
			}
			elasticsearchController.enqueueElasticsearchCluster(cur)
		},
		DeleteFunc: elasticsearchController.enqueueElasticsearchClusterDelete,
	})
	elasticsearchController.esLister = es.Lister()
	elasticsearchController.esListerSynced = es.Informer().HasSynced

	// add an event handler to the Deployment informer
	deploys.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.handleDeploy,
		UpdateFunc: func(old, cur interface{}) {
			if reflect.DeepEqual(old, cur) {
				return
			}
			elasticsearchController.handleDeploy(cur)
		},
		DeleteFunc: func(obj interface{}) {
			elasticsearchController.handleDeploy(obj)
		},
	})
	elasticsearchController.deployLister = deploys.Lister()
	elasticsearchController.deployListerSynced = deploys.Informer().HasSynced

	// add an event handler to the StatefulSet informer
	statefulsets.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.handleStatefulSet,
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			elasticsearchController.handleStatefulSet(new)
		},
		DeleteFunc: elasticsearchController.handleStatefulSet,
	})
	elasticsearchController.statefulSetLister = statefulsets.Lister()
	elasticsearchController.statefulSetListerSynced = statefulsets.Informer().HasSynced

	// add an event handler to the ServiceAccount informer
	serviceaccounts.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.handleServiceAccount,
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			elasticsearchController.handleServiceAccount(new)
		},
		DeleteFunc: elasticsearchController.handleServiceAccount,
	})
	elasticsearchController.serviceAccountLister = serviceaccounts.Lister()
	elasticsearchController.serviceAccountListerSynced = serviceaccounts.Informer().HasSynced

	// add an event handler to the Service informer
	services.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: elasticsearchController.handleService,
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			elasticsearchController.handleService(new)
		},
		DeleteFunc: elasticsearchController.handleService,
	})
	elasticsearchController.serviceLister = services.Lister()
	elasticsearchController.serviceListerSynced = services.Informer().HasSynced

	// create the actual ElasticsearchCluster controller
	elasticsearchController.elasticsearchClusterControl = NewElasticsearchClusterControl(
		elasticsearchController.statefulSetLister,
		elasticsearchController.deployLister,
		elasticsearchController.serviceAccountLister,
		elasticsearchController.serviceLister,
		NewElasticsearchClusterNodePoolControl(
			cl,
			elasticsearchController.deployLister,
			recorder,
		),
		NewStatefulElasticsearchClusterNodePoolControl(
			cl,
			elasticsearchController.statefulSetLister,
			recorder,
		),
		NewElasticsearchClusterServiceAccountControl(
			cl,
			recorder,
		),
		// client service controller
		NewElasticsearchClusterServiceControl(
			cl,
			recorder,
			ServiceControlConfig{
				NameSuffix: "clients",
				EnableHTTP: true,
				Roles:      []string{"client"},
			},
		),
		// discovery service controller
		NewElasticsearchClusterServiceControl(
			cl,
			recorder,
			ServiceControlConfig{
				NameSuffix:  "discovery",
				Annotations: map[string]string{"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true"},
			},
		),
		recorder,
	)

	return elasticsearchController
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
