package cassandra

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1beta1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	rbacinformers "k8s.io/client-go/informers/rbac/v1beta1"

	navigatorclientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navigatorinformers "github.com/jetstack/navigator/pkg/client/informers/externalversions/navigator/v1alpha1"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/role"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/rolebinding"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/seedlabeller"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/serviceaccount"
)

// NewCassandra returns a new CassandraController that can be used
// to monitor for CassandraCluster resources and create clusters in a target Kubernetes
// cluster.
//
// It accepts a list of informers that are then used to monitor the state of the
// target cluster.
type CassandraController struct {
	control                     ControlInterface
	cassLister                  listersv1alpha1.CassandraClusterLister
	statefulSetLister           appslisters.StatefulSetLister
	cassListerSynced            cache.InformerSynced
	serviceListerSynced         cache.InformerSynced
	statefulSetListerSynced     cache.InformerSynced
	pilotsListerSynced          cache.InformerSynced
	podsListerSynced            cache.InformerSynced
	serviceAccountsListerSynced cache.InformerSynced
	rolesListerSynced           cache.InformerSynced
	roleBindingsListerSynced    cache.InformerSynced
	queue                       workqueue.RateLimitingInterface
	recorder                    record.EventRecorder
}

func NewCassandra(
	naviClient navigatorclientset.Interface,
	kubeClient kubernetes.Interface,
	cassClusters navigatorinformers.CassandraClusterInformer,
	services coreinformers.ServiceInformer,
	statefulSets appsinformers.StatefulSetInformer,
	pilots navigatorinformers.PilotInformer,
	pods coreinformers.PodInformer,
	serviceAccounts coreinformers.ServiceAccountInformer,
	roles rbacinformers.RoleInformer,
	roleBindings rbacinformers.RoleBindingInformer,
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
	// add an event handler to the Pod informer
	pods.Informer().AddEventHandler(
		&controllers.BlockingEventHandler{
			WorkFunc: cc.handlePodObject,
		},
	)
	cc.cassLister = cassClusters.Lister()
	cc.statefulSetLister = statefulSets.Lister()
	cc.cassListerSynced = cassClusters.Informer().HasSynced
	cc.serviceListerSynced = services.Informer().HasSynced
	cc.statefulSetListerSynced = statefulSets.Informer().HasSynced
	cc.pilotsListerSynced = pilots.Informer().HasSynced
	cc.podsListerSynced = pods.Informer().HasSynced
	cc.serviceAccountsListerSynced = serviceAccounts.Informer().HasSynced
	cc.rolesListerSynced = roles.Informer().HasSynced
	cc.roleBindingsListerSynced = roleBindings.Informer().HasSynced
	cc.control = NewControl(
		service.NewControl(
			kubeClient,
			services.Lister(),
			recorder,
			service.NodesServiceForCluster,
		),
		service.NewControl(
			kubeClient,
			services.Lister(),
			recorder,
			service.SeedsServiceForCluster,
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
		serviceaccount.NewControl(
			kubeClient,
			serviceAccounts.Lister(),
			recorder,
		),
		role.NewControl(
			kubeClient,
			roles.Lister(),
			recorder,
		),
		rolebinding.NewControl(
			kubeClient,
			roleBindings.Lister(),
			recorder,
		),
		seedlabeller.NewControl(
			kubeClient,
			statefulSets.Lister(),
			pods.Lister(),
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
		e.statefulSetListerSynced,
		e.pilotsListerSynced,
		e.podsListerSynced,
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

func (e *CassandraController) enqueueCassandraCluster(obj interface{}) {
	key, err := controllers.KeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	glog.V(4).Infof("Adding Cassandra Cluster '%s' to queue", key)
	e.queue.AddRateLimited(key)
}

func (e *CassandraController) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		glog.Errorf("error decoding object, invalid type")
		return
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	ownerRef := metav1.GetControllerOf(object)
	if ownerRef == nil || ownerRef.Kind != "CassandraCluster" {
		return
	}

	cluster, err := e.cassLister.CassandraClusters(object.GetNamespace()).Get(ownerRef.Name)
	if err != nil {
		glog.V(4).Infof(
			"ignoring orphaned object '%s' of cassandracluster '%s'",
			object.GetSelfLink(), ownerRef.Name,
		)
		return
	}

	e.enqueueCassandraCluster(cluster)
}

// getPodOwner will return the owning ElasticsearchCluster for a pod by
// first looking up it's owning StatefulSet, and then finding the
// CassandraCluster that owns that StatefulSet. If the pod is not managed
// by a StatefulSet/CassandraCluster, it will do nothing.
func (e *CassandraController) handlePodObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		glog.Errorf("error decoding object, invalid type")
		return
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	ownerRef := metav1.GetControllerOf(object)
	if ownerRef == nil || ownerRef.Kind != "StatefulSet" {
		return
	}
	ss, err := e.statefulSetLister.StatefulSets(object.GetNamespace()).Get(ownerRef.Name)
	if err != nil {
		glog.V(4).Infof(
			"ignoring orphaned object '%s' of statefulset '%s'",
			object.GetSelfLink(), ownerRef.Name,
		)
		return
	}

	e.handleObject(ss)
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
		ctx.KubeSharedInformerFactory.Core().V1().ServiceAccounts(),
		ctx.KubeSharedInformerFactory.Rbac().V1beta1().Roles(),
		ctx.KubeSharedInformerFactory.Rbac().V1beta1().RoleBindings(),
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
