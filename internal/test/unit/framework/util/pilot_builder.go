package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

type pilotBuilder struct {
	pilot *v1alpha1.Pilot
}

func NewPilot(name, namespace string) *pilotBuilder {
	return &pilotBuilder{
		pilot: &v1alpha1.Pilot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}

func (e *pilotBuilder) SetESCluster(name string) *pilotBuilder {
	if e.pilot.Labels == nil {
		e.pilot.Labels = map[string]string{}
	}
	e.pilot.Labels[v1alpha1.ElasticsearchClusterNameLabel] = name
	return e
}

func (e *pilotBuilder) SetESNodePool(name string) *pilotBuilder {
	if e.pilot.Labels == nil {
		e.pilot.Labels = map[string]string{}
	}
	e.pilot.Labels[v1alpha1.ElasticsearchNodePoolNameLabel] = name
	return e
}
