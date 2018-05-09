package actions

import (
	"fmt"

	apps "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerror "k8s.io/apimachinery/pkg/util/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type ScaleIn struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &ScaleIn{}

func (a *ScaleIn) Name() string {
	return "ScaleIn"
}

func (a *ScaleIn) Execute(s *controllers.State) error {
	ss := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	ss, err := s.StatefulSetLister.StatefulSets(ss.Namespace).Get(ss.Name)
	if err != nil {
		return err
	}
	ss = ss.DeepCopy()
	if *ss.Spec.Replicas > *a.NodePool.Replicas {
		pilots, err := pilotsForStatefulSet(s, a.Cluster, a.NodePool, ss)
		if err != nil {
			return err
		}

		if len(pilots) <= 1 {
			return fmt.Errorf(
				"Not enough pilots to scale down: %d",
				len(pilots),
			)
		}

		allDecommissioned := true

		nPilotsToRemove := int(*ss.Spec.Replicas - *a.NodePool.Replicas)
		for i := 1; i <= nPilotsToRemove; i++ {
			p := pilots[len(pilots)-i].DeepCopy()
			if p.Spec.Cassandra == nil {
				p.Spec.Cassandra = &v1alpha1.CassandraPilotSpec{}
			}

			if !p.Spec.Cassandra.Decommissioned {
				p.Spec.Cassandra.Decommissioned = true
				_, err := s.NavigatorClientset.NavigatorV1alpha1().Pilots(p.Namespace).Update(p)
				if err != nil {
					return err
				}

				s.Recorder.Eventf(
					p,
					corev1.EventTypeNormal,
					a.Name(),
					"Marked cassandra pilot %s for decommission", p.Name,
				)
			}

			if p.Status.Cassandra == nil {
				p.Status.Cassandra = &v1alpha1.CassandraPilotStatus{}
			}

			if !p.Status.Cassandra.Decommissioned {
				allDecommissioned = false
			}
		}

		if allDecommissioned {
			ss.Spec.Replicas = a.NodePool.Replicas
			_, err = s.Clientset.AppsV1beta1().StatefulSets(ss.Namespace).Update(ss)
			if err == nil {
				s.Recorder.Eventf(
					a.Cluster,
					corev1.EventTypeNormal,
					a.Name(),
					"All cassandra nodes decommissioned, scaling cluster to size %d", a.NodePool.Replicas,
				)
			}
		}
	}
	if *ss.Spec.Replicas < *a.NodePool.Replicas {
		return fmt.Errorf(
			"the NodePool.Replicas value (%d) "+
				"is greater than the existing StatefulSet.Replicas value (%d)",
			a.NodePool.Replicas, *ss.Spec.Replicas,
		)
	}
	return nil
}

func pilotNameForStatefulSetReplica(setName string, replica int32) string {
	return fmt.Sprintf("%s-%d", setName, replica)
}

func pilotsForStatefulSet(state *controllers.State, cluster *v1alpha1.CassandraCluster, nodePool *v1alpha1.CassandraClusterNodePool, statefulSet *apps.StatefulSet) ([]*v1alpha1.Pilot, error) {
	replicasPtr := statefulSet.Spec.Replicas
	if replicasPtr == nil {
		return nil, fmt.Errorf("statefulset.spec.replicas cannot be nil")
	}
	replicas := *replicasPtr
	// TODO: read the cluster name and node pool name from the statefulset
	// metadata instead of the Scale structure so we can make this a package
	// function. For now, this way is safest until we are sure these
	// labels are going to remain stable
	selector, err := util.SelectorForNodePool(cluster, nodePool.Name)
	if err != nil {
		return nil, err
	}
	pilots, err := state.PilotLister.Pilots(cluster.Namespace).List(selector)
	if err != nil {
		return nil, err
	}
	var errs []error
	var output []*v1alpha1.Pilot
Outer:
	for i := int32(0); i < replicas; i++ {
		pilotName := pilotNameForStatefulSetReplica(statefulSet.Name, i)
		for _, p := range pilots {
			if p.Name == pilotName {
				output = append(output, p)
				continue Outer
			}
		}
		errs = append(errs, fmt.Errorf("pilot %q not found", pilotName))
	}
	if len(errs) > 0 {
		return nil, &k8sErrors.StatusError{
			ErrStatus: metav1.Status{
				Message: utilerror.NewAggregate(errs).Error(),
				Reason:  metav1.StatusReasonNotFound,
			},
		}
	}
	return output, nil
}
