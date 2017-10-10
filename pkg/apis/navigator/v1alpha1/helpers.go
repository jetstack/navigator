package v1alpha1

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (p *Pilot) HasCondition(condition PilotCondition) bool {
	if len(p.Status.Conditions) == 0 {
		return false
	}
	for _, cond := range p.Status.Conditions {
		if condition.Type == cond.Type && condition.Status == cond.Status {
			return true
		}
	}
	return false
}

func (p *Pilot) UpdateStatusCondition(conditionType PilotConditionType, status ConditionStatus, reason, message string, format ...string) {
	newCondition := PilotCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: fmt.Sprintf(message, format),
	}

	t := time.Now()

	if len(p.Status.Conditions) == 0 {
		glog.Infof("Setting lastTransitionTime for Pilot %q condition %q to %v", p.Name, conditionType, t)
		newCondition.LastTransitionTime = metav1.NewTime(t)
		p.Status.Conditions = []PilotCondition{newCondition}
	} else {
		for i, cond := range p.Status.Conditions {
			if cond.Type == conditionType {
				if cond.Status != newCondition.Status {
					glog.Infof("Found status change for Pilot %q condition %q: %q -> %q; setting lastTransitionTime to %v", p.Name, conditionType, cond.Status, status, t)
					newCondition.LastTransitionTime = metav1.NewTime(t)
				} else {
					newCondition.LastTransitionTime = cond.LastTransitionTime
				}

				p.Status.Conditions[i] = newCondition
				break
			}
		}
	}
}
