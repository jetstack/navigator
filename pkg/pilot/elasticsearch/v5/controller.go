package v5

import (
	"context"
	"fmt"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	elasticsearchLocalNodeID = "_local"
)

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	// TODO: perform cluster wide actions if we are a leader
	if !p.genericPilot.IsThisPilot(pilot) {
		return nil
	}
	if pilot.Status.Elasticsearch == nil {
		pilot.Status.Elasticsearch = &v1alpha1.ElasticsearchPilotStatus{}
	}
	// set the ES document count to nil to indicate an unknown number of
	// documents in case obtaining the document count fails. This prevents the
	// node being shut down if it is not safe to do so.
	pilot.Status.Elasticsearch.Documents = nil
	if p.localESClient == nil {
		return fmt.Errorf("local elasticsearch client not available")
	}
	// TODO: use context with a timeout
	statsList, err := p.localESClient.NodesStats().NodeId(elasticsearchLocalNodeID).Do(context.Background())
	if err != nil {
		return err
	}
	// we can iterate over the results as the elastic client should only return
	// a single node entry or none as we specify a single nodeID.
	for _, stats := range statsList.Nodes {
		docCount := stats.Indices.Docs.Count
		pilot.Status.Elasticsearch.Documents = &docCount
	}
	return nil
}
