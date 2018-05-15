package resources

import (
	apiv1 "k8s.io/api/core/v1"
)

func RequirementsEqual(a, b apiv1.ResourceRequirements) bool {
	if a.Limits.Cpu().Cmp(*b.Limits.Cpu()) != 0 {
		return false
	}

	if a.Limits.Memory().Cmp(*b.Limits.Memory()) != 0 {
		return false
	}

	if a.Requests.Cpu().Cmp(*b.Requests.Cpu()) != 0 {
		return false
	}

	if a.Requests.Memory().Cmp(*b.Requests.Memory()) != 0 {
		return false
	}

	return true
}
