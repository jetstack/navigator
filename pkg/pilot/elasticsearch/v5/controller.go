package v5

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	elasticsearchLocalNodeID = "_local"
	nodeStatsUpdateFrequency = time.Second * 10
)

func (p *Pilot) syncFunc(pilot *v1alpha1.Pilot) error {
	glog.V(4).Infof("ElasticsearchController: syncing current pilot %q", pilot.Name)
	if pilot.Status.Elasticsearch == nil {
		pilot.Status.Elasticsearch = &v1alpha1.ElasticsearchPilotStatus{}
	}

	// if true, it is too soon to perform an update of the node stats
	if p.lastNodeStatsUpdate.Add(nodeStatsUpdateFrequency).After(time.Now()) {
		return nil
	}

	if err := p.updateNodeStats(pilot); err != nil {
		return err
	}
	if err := p.updateNodeInfo(pilot); err != nil {
		return err
	}
	return nil
}

func (p *Pilot) updateNodeStats(pilot *v1alpha1.Pilot) error {
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
	glog.V(4).Infof("Got %d nodes in returned data", len(statsList.Nodes))
	// we can iterate over the results as the elastic client should only return
	// a single node entry or none as we specify a single nodeID.
	for name, stats := range statsList.Nodes {
		glog.V(4).Infof("Applying stats for node %q", name)
		docCount := stats.Indices.Docs.Count
		pilot.Status.Elasticsearch.Documents = &docCount
	}
	return nil
}
func (p *Pilot) updateNodeInfo(pilot *v1alpha1.Pilot) error {
	if pilot.Status.Elasticsearch == nil {
		pilot.Status.Elasticsearch = &v1alpha1.ElasticsearchPilotStatus{}
	}
	if p.localESClient == nil {
		return fmt.Errorf("local elasticsearch client not available")
	}
	// TODO: use context with a timeout
	infoList, err := p.localESClient.NodesInfo().NodeId(p.Options.GenericPilotOptions.PilotName).Do(context.Background())
	if err != nil {
		return err
	}
	glog.V(4).Infof("Got %d nodes in returned data", len(infoList.Nodes))
	// we can iterate over the results as the elastic client should only return
	// a single node entry or none as we specify a single nodeID.
	for name, info := range infoList.Nodes {
		glog.V(4).Infof("Applying info for node %q", name)
		pilot.Status.Elasticsearch.Version = info.Version
	}
	return nil
}

// Perform leader elected sync actions. A broad description of this controllers
// behaviour:
//
// - (TODO) Generate the exclude allocation elasticsearch string
//   - List all Pilots in cluster
//   - Filter Pilots to those that have excludeShards: true
//   - Build exclude string
//   - Update ES API
// - Update ElasticsearchCluster.status field
//   - Check the current cluster health status in the Elasticsearch API
//   - Update the status block
//   - Write changes back to k8s API
func (p *Pilot) leaderElectedSyncFunc(pilot *v1alpha1.Pilot) error {
	glog.V(4).Infof("ElasticsearchController: leader elected sync of pilot %q", pilot.Name)
	ctx := context.Background()

	// TODO: switch on currentHealth.Status to translate into correct concrete types
	esc, err := p.getOwningCluster(pilot)
	if err != nil {
		return err
	}

	err = p.updateElasticsearchClusterStatus(ctx, esc)
	if err != nil {
		return fmt.Errorf("error updating elasticsearchcluster %q status: %s", esc.Name, err)
	}

	return nil
}

func (p *Pilot) getOwningCluster(pilot *v1alpha1.Pilot) (*v1alpha1.ElasticsearchCluster, error) {
	ownerRef := metav1.GetControllerOf(pilot)
	if ownerRef == nil || ownerRef.Kind != "ElasticsearchCluster" {
		return nil, fmt.Errorf("could not determine controller of Pilot %q", pilot.Name)
	}

	return p.esClusterLister.ElasticsearchClusters(pilot.Namespace).Get(ownerRef.Name)
}

func (p *Pilot) updateElasticsearchClusterStatus(ctx context.Context, esc *v1alpha1.ElasticsearchCluster) error {
	// TODO: use a cluster-wide elasticsearch client instead of localESClient
	if p.localESClient == nil {
		return fmt.Errorf("local elasticsearch client not available")
	}
	currentHealth, err := p.localESClient.ClusterHealth().Do(ctx)
	if err != nil {
		return err
	}

	escCopy := esc.DeepCopy()
	escCopy.Status.Health = parseHealth(currentHealth.Status)
	if _, err := p.navigatorClient.NavigatorV1alpha1().ElasticsearchClusters(escCopy.Namespace).UpdateStatus(escCopy); err != nil {
		return err
	}

	return nil
}

func parseHealth(s string) v1alpha1.ElasticsearchClusterHealth {
	switch strings.ToLower(s) {
	case "green":
		return v1alpha1.ElasticsearchClusterHealthGreen
	case "yellow":
		return v1alpha1.ElasticsearchClusterHealthYellow
	case "red":
		return v1alpha1.ElasticsearchClusterHealthRed
	}
	return v1alpha1.ElasticsearchClusterHealth(s)
}
