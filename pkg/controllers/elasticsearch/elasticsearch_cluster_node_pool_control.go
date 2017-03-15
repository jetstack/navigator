package elasticsearch

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/api"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/record"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
)

type ElasticsearchClusterNodePoolControl interface {
	SyncElasticsearchClusterNodePool(*v1.ElasticsearchCluster, *v1.ElasticsearchClusterNodePool) error
}

type defaultElasticsearchClusterNodePoolControl struct {
	kubeClient        *kubernetes.Clientset
	statefulSetLister appslisters.StatefulSetLister
	deploymentLister  extensionslisters.DeploymentLister

	recorder record.EventRecorder
}

var _ ElasticsearchClusterNodePoolControl = &defaultElasticsearchClusterNodePoolControl{}

func NewElasticsearchClusterNodePoolControl(
	kubeClient *kubernetes.Clientset,
	statefulSetLister appslisters.StatefulSetLister,
	deploymentLister extensionslisters.DeploymentLister,
	recorder record.EventRecorder,
) ElasticsearchClusterNodePoolControl {
	return &defaultElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		deploymentLister:  deploymentLister,
		recorder:          recorder,
	}
}

func (e *defaultElasticsearchClusterNodePoolControl) SyncElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	if np.State != nil && np.State.Stateful {
		return e.syncStatefulNodePool(c, np)
	}
	return e.syncNodePool(c, np)
}

func (e *defaultElasticsearchClusterNodePoolControl) syncStatefulNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	ss, err := e.statefulSetLister.StatefulSets(c.Namespace).Get(nodePoolResourceName(c, np))

	if errors.IsNotFound(err) {
		err = e.createStatefulSet(c, np)
		e.recordNodePoolEvent("create", c, np, err)
		return err
	}

	if err != nil {
		e.recordNodePoolEvent("sync", c, np, err)
		return err
	}

	// // Disabled for now
	// if e.statefulSetNeedsRecreating(c, np, ss) {
	// 	// TODO: write logic to delete and recreate the statefulset
	// }
	if copy, err := api.Scheme.DeepCopy(ss); err == nil {
		ss = copy.(*apps.StatefulSet)
	} else {
		e.recordNodePoolEvent("sync", c, np, err)
		utilruntime.HandleError(fmt.Errorf("error copying updated Pod: %v", err))
	}

	// if np.Replicas != int64(*ss.Spec.Replicas) {
	// 	e.kubeClient.Apps().StatefulSets(ss.Namespace)
	// }

	return nil
}

func (e *defaultElasticsearchClusterNodePoolControl) syncNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	depl, err := e.deploymentLister.Deployments(c.Namespace).Get(nodePoolResourceName(c, np))

	if errors.IsNotFound(err) {
		err = e.createDeployment(c, np)
		e.recordNodePoolEvent("create", c, np, err)
		return err
	}

	if err != nil {
		e.recordNodePoolEvent("sync", c, np, err)
		return err
	}

	// // Disabled for now
	// if e.statefulSetNeedsRecreating(c, np, ss) {
	// 	// TODO: write logic to delete and recreate the statefulset
	// }
	if copy, err := api.Scheme.DeepCopy(depl); err == nil {
		depl = copy.(*extensions.Deployment)
	} else {
		e.recordNodePoolEvent("sync", c, np, err)
		utilruntime.HandleError(fmt.Errorf("error copying updated Pod: %v", err))
	}

	// if np.Replicas != int64(*ss.Spec.Replicas) {
	// 	e.kubeClient.Apps().StatefulSets(ss.Namespace)
	// }

	return nil
}

func (e *defaultElasticsearchClusterNodePoolControl) createDeployment(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	elasticsearchPodTemplate, err := elasticsearchPodTemplateSpec(c, np)

	if err != nil {
		return fmt.Errorf("error building elasticsearch container: %s", err.Error())
	}

	depl := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nodePoolResourceName(c, np),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Annotations: map[string]string{
				nodePoolVersionAnnotationKey: c.Spec.Version,
			},
			Labels: buildNodePoolLabels(c, np.Name, np.Roles...),
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32Ptr(int32(np.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: buildNodePoolLabels(c, np.Name, np.Roles...),
			},
			Template: *elasticsearchPodTemplate,
		},
	}

	depl, err = e.kubeClient.Extensions().Deployments(c.Namespace).Create(depl)

	if err != nil {
		return fmt.Errorf("error creating deployment: %s", err.Error())
	}

	return nil
}

func (e *defaultElasticsearchClusterNodePoolControl) createStatefulSet(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	volumeClaimTemplateAnnotations,
		volumeResourceRequests :=
		map[string]string{},
		apiv1.ResourceList{}

	if np.State.Persistence != nil {
		if np.State.Persistence.StorageClass != "" {
			volumeClaimTemplateAnnotations["volume.beta.kubernetes.io/storage-class"] = np.State.Persistence.StorageClass
		}

		if size := np.State.Persistence.Size; size != "" {
			storageRequests, err := resource.ParseQuantity(size)

			if err != nil {
				return fmt.Errorf("error parsing storage size quantity '%s': %s", size, err.Error())
			}

			volumeResourceRequests[apiv1.ResourceStorage] = storageRequests
		}
	}

	elasticsearchPodTemplate, err := elasticsearchPodTemplateSpec(c, np)

	if err != nil {
		return fmt.Errorf("error building elasticsearch container: %s", err.Error())
	}

	ss := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nodePoolResourceName(c, np),
			Namespace:       c.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference(c)},
			Annotations: map[string]string{
				nodePoolVersionAnnotationKey: c.Spec.Version,
			},
			Labels: buildNodePoolLabels(c, np.Name, np.Roles...),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    int32Ptr(int32(np.Replicas)),
			ServiceName: nodePoolResourceName(c, np),
			Selector: &metav1.LabelSelector{
				MatchLabels: buildNodePoolLabels(c, np.Name, np.Roles...),
			},
			Template: *elasticsearchPodTemplate,
			VolumeClaimTemplates: []apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "elasticsearch-data",
						Annotations: volumeClaimTemplateAnnotations,
					},
					Spec: apiv1.PersistentVolumeClaimSpec{
						AccessModes: []apiv1.PersistentVolumeAccessMode{
							apiv1.ReadWriteOnce,
						},
						Resources: apiv1.ResourceRequirements{
							Requests: volumeResourceRequests,
						},
					},
				},
			},
		},
	}

	ss, err = e.kubeClient.Apps().StatefulSets(c.Namespace).Create(ss)

	if err != nil {
		return fmt.Errorf("error creating statefulset: %s", err.Error())
	}

	return nil
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterNodePoolControl) recordNodePoolEvent(verb string, cluster *v1.ElasticsearchCluster, pool *v1.ElasticsearchClusterNodePool, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s NodePool %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), pool.Name, cluster.Name)
		e.recorder.Event(cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s NodePool %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), pool.Name, cluster.Name, err)
		e.recorder.Event(cluster, apiv1.EventTypeWarning, reason, message)
	}
}
