package elastic

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
	"gitlab.jetstack.net/marshal/colonel/pkg/util/errors"
)

const (
	typeName = "es"
	kindName = "ElasticsearchCluster"

	nodePoolVersionAnnotationKey = "elasticsearch.marshal.io/deployed-version"
)

var (
	trueVar = true
)

func (e *ElasticsearchController) nodePoolNeedsUpdate(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (bool, error) {
	if np.State.Stateful {
		return e.statefulNodePoolNeedsUpdate(c, np)
	}

	return false, nil
}

func (e *ElasticsearchController) deploymentNodePoolNeedsUpdate(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (bool, error) {
	if np.State.Stateful {
		return false, fmt.Errorf("node pool is stateful, but deploymentNodePoolNeedsUpdate called")
	}

	nodePoolName := nodePoolResourceName(c, np)
	depl, err := e.deployLister.Deployments(c.Namespace).Get(nodePoolName)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return true, nil
		}

		return false, errors.Transient(fmt.Errorf("error getting deployment '%s' from apiserver: %s", nodePoolName, err.Error()))
	}

	// if this deployment is not marked as managed by the cluster, exit with an error and not performing an update to prevent
	// standing on the cluster administrators toes
	if !isManagedByCluster(c, depl.ObjectMeta) {
		return false, fmt.Errorf("found existing deployment with name, but it is not owned by this ElasticsearchCluster. not updating!")
	}

	// if the desired number of replicas is not equal to the actual
	if *depl.Spec.Replicas != int32(np.Replicas) {
		return true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if nodePoolVersionAnnotation(depl.Annotations) != c.Spec.Version {
		return true, nil
	}

	return false, nil
}

func (e *ElasticsearchController) statefulNodePoolNeedsUpdate(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) (bool, error) {
	if !np.State.Stateful {
		return false, fmt.Errorf("node pool is not stateful, but statefulNodePoolNeedsUpdate called")
	}

	nodePoolName := nodePoolResourceName(c, np)
	ss, err := e.statefulSetLister.StatefulSets(c.Namespace).Get(nodePoolName)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return true, nil
		}

		return false, errors.Transient(fmt.Errorf("error getting statefulset '%s' from apiserver: %s", nodePoolName, err.Error()))
	}

	// if this statefulset is not marked as managed by the cluster, exit with an error and not performing an update to prevent
	// standing on the cluster administrators toes
	if !isManagedByCluster(c, ss.ObjectMeta) {
		return false, fmt.Errorf("found existing statefulset with name, but it is not owned by this ElasticsearchCluster. not updating!")
	}

	// if the desired number of replicas is not equal to the actual
	if *ss.Spec.Replicas != int32(np.Replicas) {
		return true, nil
	}

	// if the version of the cluster has changed, trigger an update
	if nodePoolVersionAnnotation(ss.Annotations) != c.Spec.Version {
		return true, nil
	}

	return false, nil
	// container, ok := ss.Spec.Template.Spec.Containers[0]

	// // somehow there are no containers in this Pod - trigger an update
	// if !ok {
	// 	return true, nil
	// }

	// if
}

func isManagedByCluster(c *v1.ElasticsearchCluster, meta metav1.ObjectMeta) bool {
	clusterOwnerRef := ownerReference(c)
	for _, o := range meta.OwnerReferences {
		if o == clusterOwnerRef {
			return true
		}
	}
	return false
}

func ownerReference(c *v1.ElasticsearchCluster) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: c.APIVersion,
		Kind:       c.Kind,
		Name:       c.Name,
		UID:        c.UID,
		Controller: &trueVar,
	}
}

func nodePoolResourceName(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) string {
	return fmt.Sprintf("%s-%s-%s", typeName, c.Name, np.Name)
}

func nodePoolVersionAnnotation(m map[string]string) string {
	return m[nodePoolVersionAnnotationKey]
}
