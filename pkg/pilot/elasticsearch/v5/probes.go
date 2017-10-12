package v5

import (
	"context"
	"time"

	"github.com/golang/glog"

	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/probe"
)

func (p *Pilot) ReadinessCheck() probe.Check {
	return probe.CombineChecks(
		p.localNodeHealth,
	)
}

func (p *Pilot) LivenessCheck() probe.Check {
	return probe.CombineChecks(
		func() error {
			return nil
		},
	)
}

// Check the health of this Elasticsearch node
func (p *Pilot) localNodeHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := p.localESClient.ClusterHealth().Local(true).Do(ctx)
	if err != nil {
		return err
	}
	glog.V(2).Infof("Local node health status: %q", resp.Status)
	return nil
}
