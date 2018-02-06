package actions

import (
	"fmt"

	core "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

type CreatePilot struct {
	Cluster  *v1alpha1.ElasticsearchCluster
	NodePool *v1alpha1.ElasticsearchClusterNodePool
}

var _ controllers.Action = &CreatePilot{}

func (c *CreatePilot) Name() string {
	return "CreatePilot"
}

func (c *CreatePilot) Message() string {
	return fmt.Sprintf("Created Pilot resources for node pool %q", c.NodePool.Name)
}

func (c *CreatePilot) Execute(state *controllers.State) error {
	selector, err := util.SelectorForNodePool(c.Cluster.Name, c.NodePool.Name)
	if err != nil {
		return err
	}

	allPods, err := state.PodLister.Pods(c.Cluster.Namespace).List(selector)
	if err != nil {
		return err
	}

	for _, pod := range allPods {
		isMember, err := controllers.PodControlledByCluster(c.Cluster, pod, state.StatefulSetLister)
		if err != nil {
			return fmt.Errorf("error checking if pod is controller by elasticsearch cluster: %s", err.Error())
		}

		// skip this pod if it's not a member of the cluster
		if !isMember {
			continue
		}

		pilot := newPilotResource(c.Cluster, pod)
		_, err = state.NavigatorClientset.NavigatorV1alpha1().Pilots(pilot.Namespace).Create(pilot)
		if k8sErrors.IsAlreadyExists(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error ensuring pilot resource exists for pod '%s': %s", pod.Name, err.Error())
		}
		state.Recorder.Eventf(c.Cluster, core.EventTypeNormal, c.Name(), "Created pilot %q", pilot.Name)
	}
	return nil
}

func newPilotResource(c *v1alpha1.ElasticsearchCluster, pod *core.Pod) *v1alpha1.Pilot {
	// TODO: break this function out to account for scale down events, and
	// setting the spec however appropriate
	pilot := &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(c)},
			Labels:          pod.Labels,
		},
		Spec: v1alpha1.PilotSpec{
			Elasticsearch: &v1alpha1.PilotElasticsearchSpec{},
		},
	}
	return pilot
}
