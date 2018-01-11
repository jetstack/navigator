package v5

import (
	"context"
	"fmt"

	"github.com/golang/glog"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	elasticsearchLocalNodeID = "_local"
)

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	glog.V(4).Infof("ElasticsearchController: syncing current pilot %q", pilot.Name)
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
	statsList, err := p.localESClient.NodesStats().NodeId(p.Options.GenericPilotOptions.PilotName).Do(context.Background())
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

func (p *Pilot) leaderElectedSyncFunc(pilot *v1alpha1.Pilot) error {
	glog.V(4).Infof("ElasticsearchController: leader elected sync of pilot %q", pilot.Name)
	return nil
}
