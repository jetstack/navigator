package elasticsearch

import (
	"fmt"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listers "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/actions"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/configmap"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/role"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/rolebinding"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/service"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/serviceaccount"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

const (
	errorSync = "ErrSync"

	messageErrorSyncServiceAccount = "Error syncing service account: %s"
	messageErrorSyncConfigMap      = "Error syncing config map: %s"
	messageErrorSyncService        = "Error syncing service: %s"
	messageErrorSyncNodePools      = "Error syncing node pools: %s"
	messageErrorSyncRoles          = "Error syncing RBAC roles: %s"
	messageErrorSyncRoleBindings   = "Error syncing RBAC role bindings: %s"
	messageSuccessExecuteAction    = "Successfully executed action"
)

type ControlInterface interface {
	Sync(*v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error)
}

var _ ControlInterface = &defaultElasticsearchClusterControl{}

type defaultElasticsearchClusterControl struct {
	kubeClient      kubernetes.Interface
	navigatorClient clientset.Interface

	statefulSetLister    appslisters.StatefulSetLister
	serviceAccountLister corelisters.ServiceAccountLister
	serviceLister        corelisters.ServiceLister
	configMapLister      corelisters.ConfigMapLister
	pilotLister          listers.PilotLister
	podLister            corelisters.PodLister

	configMapControl      configmap.Interface
	serviceAccountControl serviceaccount.Interface
	serviceControl        service.Interface
	roleControl           role.Interface
	roleBindingControl    rolebinding.Interface

	recorder record.EventRecorder
}

var _ ControlInterface = &defaultElasticsearchClusterControl{}

func NewController(
	kubeClient kubernetes.Interface,
	navigatorClient clientset.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	serviceAccountLister corelisters.ServiceAccountLister,
	serviceLister corelisters.ServiceLister,
	configMapLister corelisters.ConfigMapLister,
	pilotLister listers.PilotLister,
	podLister corelisters.PodLister,
	configMapControl configmap.Interface,
	serviceAccountControl serviceaccount.Interface,
	serviceControl service.Interface,
	roleControl role.Interface,
	roleBindingControl rolebinding.Interface,
	recorder record.EventRecorder,
) ControlInterface {
	return &defaultElasticsearchClusterControl{
		kubeClient:            kubeClient,
		navigatorClient:       navigatorClient,
		statefulSetLister:     statefulSetLister,
		serviceAccountLister:  serviceAccountLister,
		serviceLister:         serviceLister,
		configMapLister:       configMapLister,
		pilotLister:           pilotLister,
		podLister:             podLister,
		configMapControl:      configMapControl,
		serviceAccountControl: serviceAccountControl,
		serviceControl:        serviceControl,
		roleControl:           roleControl,
		roleBindingControl:    roleBindingControl,
		recorder:              recorder,
	}
}

func (e *defaultElasticsearchClusterControl) Sync(c *v1alpha1.ElasticsearchCluster) (v1alpha1.ElasticsearchClusterStatus, error) {
	c = c.DeepCopy()
	var err error

	defer func() {
		var errs = []error{err}
		if err := e.updateClusterStatus(c); err != nil {
			errs = append(errs, err)
		}
		err = utilerrors.NewAggregate(errs)
	}()

	if _, err = e.serviceAccountControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncServiceAccount, err.Error())
		return c.Status, err
	}

	// TODO: handle status
	if _, err = e.serviceControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncService, err.Error())
		return c.Status, err
	}

	// TODO: handle status
	if err = e.roleControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncRoles, err.Error())
		return c.Status, err
	}

	// TODO: handle status
	if err = e.roleBindingControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncRoleBindings, err.Error())
		return c.Status, err
	}

	// TODO: handle status
	if c.Status, err = e.configMapControl.Sync(c); err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncConfigMap, err.Error())
		return c.Status, err
	}

	state := &controllers.State{
		Clientset:          e.kubeClient,
		NavigatorClientset: e.navigatorClient,
		Recorder:           e.recorder,
		StatefulSetLister:  e.statefulSetLister,
		ConfigMapLister:    e.configMapLister,
		PilotLister:        e.pilotLister,
		PodLister:          e.podLister,
	}

	// for each node pool
	// - check if ConfigMap is up to date
	//   - if not, we need to create one
	//     - create configmap
	//     - update ElasticsearchCluster.status.nodePool[].desiredConfigMap with current configmap name
	// - check if statefulset uses correct configmap
	//   - if not, update statefulset with configmap name
	//     - if desiredStatefulSetHash != current hash for statefulset,
	//     - update ElasticsearchCluster.status.nodePool[].desiredStatefulSetHash with new statefulset hash from status
	nextAction, err := e.nextAction(c)
	if err != nil {
		e.recorder.Eventf(c, apiv1.EventTypeWarning, errorSync, messageErrorSyncNodePools, err.Error())
		return c.Status, err
	}

	if nextAction == nil {
		err = e.reconcileNodePools(c)
		return c.Status, err
	}

	if nextAction != nil {
		glog.Infof("Executing action %q", nextAction.Name())
		err := nextAction.Execute(state)
		glog.Infof("Finished executing action %q", nextAction.Name())
		if err != nil {
			e.recorder.Eventf(c, apiv1.EventTypeWarning, "Err"+nextAction.Name(), messageErrorSyncNodePools, err.Error())
			return c.Status, err
		}
	}

	return c.Status, err
}

func (e *defaultElasticsearchClusterControl) nextAction(c *v1alpha1.ElasticsearchCluster) (controllers.Action, error) {
	for _, np := range c.Spec.NodePools {
		action, err := e.nextActionForNodePool(c, &np)
		if err != nil {
			return nil, err
		}
		if action != nil {
			return action, nil
		}
	}
	return nil, nil
}

func (e *defaultElasticsearchClusterControl) nextActionForNodePool(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (controllers.Action, error) {
	statefulSetName := util.NodePoolResourceName(c, np)
	statefulSet, err := e.statefulSetLister.StatefulSets(c.Namespace).Get(statefulSetName)
	// create the node pool if it does not exist
	if k8sErrors.IsNotFound(err) {
		return &actions.CreateNodePool{c, np}, nil
	}
	if err != nil {
		return nil, err
	}

	// check if pilots for this statefulset are up to date
	needsUpdate, err := e.pilotsNeedUpdateForNodePool(c, np)
	if err != nil {
		return nil, err
	}
	if needsUpdate {
		return &actions.CreatePilot{c, np}, nil
	}

	currentVersionStr, ok := statefulSet.Annotations[v1alpha1.ElasticsearchNodePoolVersionAnnotation]
	if !ok {
		return nil, fmt.Errorf("cannot determine existing Elasticsearch version of statefulset %q", statefulSet.Name)
	}
	if c.Spec.Version.String() != currentVersionStr {
		return &actions.UpdateVersion{Cluster: c, NodePool: np}, nil
	}

	currentDesiredReplicas := statefulSet.Spec.Replicas
	if currentDesiredReplicas == nil {
		return nil, fmt.Errorf("current number of replicas on statefulset cannot be nil")
	}
	// if the current number of desired replicas on the statefulset does
	// not equal the number on the node pool, we need to scale
	if *currentDesiredReplicas != int32(np.Replicas) {
		return &actions.Scale{c, np, int32(np.Replicas)}, nil
	}

	return nil, nil
}

func (e *defaultElasticsearchClusterControl) updateClusterStatus(c *v1alpha1.ElasticsearchCluster) error {
	if c.Status.NodePools == nil {
		c.Status.NodePools = map[string]v1alpha1.ElasticsearchClusterNodePoolStatus{}
	}
	for _, pool := range c.Spec.NodePools {
		statefulSetName := util.NodePoolResourceName(c, &pool)
		statefulSet, err := e.statefulSetLister.StatefulSets(c.Namespace).Get(statefulSetName)
		if k8sErrors.IsNotFound(err) {
			// don't return an error if the statefulset doesn't exist
			continue
		}
		if err != nil {
			return err
		}
		poolStatus := c.Status.NodePools[pool.Name]
		poolStatus.ReadyReplicas = statefulSet.Status.ReadyReplicas
		c.Status.NodePools[pool.Name] = poolStatus
	}
	return nil
}

func (e *defaultElasticsearchClusterControl) pilotsNeedUpdateForNodePool(c *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) (bool, error) {
	selector, err := util.SelectorForNodePool(c.Name, np.Name)
	if err != nil {
		return false, err
	}

	allPods, err := e.podLister.Pods(c.Namespace).List(selector)
	if err != nil {
		return false, err
	}

	for _, pod := range allPods {
		isMember, err := controllers.PodControlledByCluster(c, pod, e.statefulSetLister)
		if err != nil {
			return false, fmt.Errorf("error checking if pod is controller by elasticsearch cluster: %s", err.Error())
		}

		// skip this pod if it's not a member of the cluster
		if !isMember {
			continue
		}

		_, err = e.pilotLister.Pilots(pod.Namespace).Get(pod.Name)
		if k8sErrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

// reconcileNodePools will look up all node pools that are owned by this
// ElasticsearchCluster resource, and delete any that are no longer referenced.
// This is used to delete old node pools that no longer exist in the cluster
// specification.
func (e *defaultElasticsearchClusterControl) reconcileNodePools(c *v1alpha1.ElasticsearchCluster) error {
	// list all statefulsets that match the clusters selector
	// loop through each node pool in c
	sel, err := util.SelectorForCluster(c.Name)
	if err != nil {
		return fmt.Errorf("error creating label selector for cluster '%s': %s", c.Name, err.Error())
	}
	sets, err := e.statefulSetLister.StatefulSets(c.Namespace).List(sel)
	if err != nil {
		return err
	}
	// we delete each statefulset that has the node pool name set to the name
	// of a valid node pool for sets
	for _, np := range c.Spec.NodePools {
		for i, ss := range sets {
			if ss.Labels != nil && ss.Labels[v1alpha1.ElasticsearchNodePoolNameLabel] == np.Name {
				sets = append(sets[:i], sets[i+1:]...)
				break
			}
		}
	}

	// delete remaining statefulsets in sets
	for _, ss := range sets {
		if !metav1.IsControlledBy(ss, c) {
			continue
		}
		err := e.kubeClient.AppsV1beta1().StatefulSets(ss.Namespace).Delete(ss.Name, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
