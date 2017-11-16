package v5

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	if p.genericPilot.IsThisPilot(pilot) {
	}
	return nil
}
