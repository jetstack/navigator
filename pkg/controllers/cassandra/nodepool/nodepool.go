package nodepool

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta2"
	"k8s.io/client-go/tools/record"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterNodepoolControl struct {
	kubeClient        kubernetes.Interface
	statefulsetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultCassandraClusterNodepoolControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulsetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterNodepoolControl{
		kubeClient:        kubeClient,
		statefulsetLister: statefulsetLister,
		recorder:          recorder,
	}
}

func (e *defaultCassandraClusterNodepoolControl) removeUnusedStatefulSets(
	cluster *v1alpha1.CassandraCluster,
) error {
	expectedStatefulSetNames := map[string]bool{}
	for _, pool := range cluster.Spec.NodePools {
		name := util.NodePoolResourceName(cluster, &pool)
		expectedStatefulSetNames[name] = true
	}
	client := e.kubeClient.AppsV1beta2().StatefulSets(cluster.Namespace)
	selector, err := util.SelectorForCluster(cluster)
	if err != nil {
		return err
	}
	existingSets, err := e.statefulsetLister.
		StatefulSets(cluster.Namespace).
		List(selector)
	if err != nil {
		return err
	}
	for _, set := range existingSets {
		if !metav1.IsControlledBy(set, cluster) {
			ownerRef := metav1.GetControllerOf(set)
			glog.Errorf(
				"Foreign owned StatefulSet: "+
					"A StatefulSet with name '%s/%s' already exists, "+
					"but it is controlled by '%v', not '%s/%s'.",
				set.Namespace, set.Name, ownerRef,
				cluster.Namespace, cluster.Name,
			)
			continue
		}
		_, found := expectedStatefulSetNames[set.Name]
		if !found {
			err := client.Delete(set.Name, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *defaultCassandraClusterNodepoolControl) createAndUpdateStatefulSets(
	cluster *v1alpha1.CassandraCluster,
) error {
	client := e.kubeClient.AppsV1beta2().StatefulSets(cluster.Namespace)
	for _, pool := range cluster.Spec.NodePools {
		glog.V(4).Infof("syncing nodepool: %#v", pool)
		desiredSet := StatefulSetForCluster(cluster, &pool)
		_, err := client.Update(desiredSet)
		if k8sErrors.IsNotFound(err) {
			_, err := client.Create(desiredSet)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	var err error
	err = e.createAndUpdateStatefulSets(cluster)
	if err != nil {
		return err
	}
	err = e.removeUnusedStatefulSets(cluster)
	if err != nil {
		return err
	}
	return err
}
