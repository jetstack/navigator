package v5

import (
	"fmt"
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	return nil
}

func (p *Pilot) leaderSyncFunc(pilot *v1alpha1.Pilot) error {
	if pilot.Spec.Elasticsearch == nil {
		runtime.HandleError(fmt.Errorf("pilot '%s' is not an Elasticsearch pilot - skipping", pilot.Name))
		return nil
	}
	return nil
	// TODO: set the exclude parameter based on current state of pilots
}
