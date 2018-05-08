package nodetool

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/jetstack/navigator/pkg/api/version"
	"github.com/jetstack/navigator/pkg/cassandra/nodetool/client"
)

// NodeState represents the cluster membership state of a C* node.
type NodeState string

const (
	NodeStateUnknown NodeState = "Unknown"
	NodeStateNormal  NodeState = "Normal"
	NodeStateLeaving NodeState = "Leaving"
	NodeStateJoining NodeState = "Joining"
	NodeStateMoving  NodeState = "Moving"
)

// NodeState represents the reachability of a C* node from the
// perspective of the node answering the query.
type NodeStatus string

const (
	NodeStatusUnknown NodeStatus = "Unknown"
	NodeStatusUp      NodeStatus = "Up"
	NodeStatusDown    NodeStatus = "Down"
)

type Node struct {
	Host   string
	ID     uuid.UUID
	State  NodeState
	Status NodeStatus
	Local  bool
}

type NodeMap map[string]*Node

func (nm NodeMap) LocalNode() *Node {
	for _, node := range nm {
		if node.Local {
			return node
		}
	}
	return nil
}

type Interface interface {
	Status() (NodeMap, error)
	Version() (*version.Version, error)
}

type tool struct {
	client client.Interface
}

var _ Interface = &tool{}

func New(client client.Interface) Interface {
	return &tool{
		client: client,
	}
}

func NewFromURL(u *url.URL) Interface {
	return New(client.New(u, &http.Client{}))
}

func setsIntersect(setsToCheck ...sets.String) bool {
	all := sets.NewString()
	totalLength := 0
	for _, s := range setsToCheck {
		totalLength += s.Len()
		all = all.Union(s)
	}
	return all.Len() != totalLength
}

// Status generates a summary of the status of every C* node in the cluster.
// From the perspective of the local node.
//
// It is intended to produce identical information to the `nodetool status` utility.
// But it only a reports a subset of the `nodetool status` information, for now.
// Enough to allow Navigator to determine whether a node is Up and Normal (healthy).
// This function returns structured information about the Cassandra cluster health,
// which avoids having to parse the unstructured, human readable output of `nodetool status`.
// Here is an example of the parsing that we are seeking to avoid:
// https://github.com/kubernetes/examples/blob/b86c9d50be45eaf5ce74dee7159ce38b0e149d38/cassandra/image/files/ready-probe.sh
// And here is the source code for the `nodetool status`:
// https://github.com/apache/cassandra/blob/cassandra-3.11.2/src/java/org/apache/cassandra/tools/nodetool/Status.java
//
// # Algorithm
//
// For every C* node that has reported its `host_id` (i.e. present in HostIdMap):
// * Determine the status of the node (one of live, unreachable, unknown)
// * Determine the state of the node (one of leaving, joining, moving, normal)
//
// We perform additional assertions to check that a node is only present in one status and one state.
// If these assertions fail, we return an error so as to avoid reporting false positive status.
// Note: `nodetool status` does not perform these assertions.
func (t *tool) Status() (NodeMap, error) {
	ssInfo, err := t.client.StorageService()
	if err != nil {
		return nil, err
	}

	nodes := NodeMap{}
	for host, id := range ssInfo.HostIdMap {
		nodes[host] = &Node{
			Host:   host,
			ID:     id,
			Status: NodeStatusUnknown,
			State:  NodeStateNormal,
		}
	}

	liveNodes := sets.NewString(ssInfo.LiveNodes...)
	unreachableNodes := sets.NewString(ssInfo.UnreachableNodes...)
	// Assert that a nodes are only in one state.
	if setsIntersect(liveNodes, unreachableNodes) {
		return nil, fmt.Errorf(
			"unexpected state: some nodes were reported in more than one status. "+
				"Live: %v, "+
				"Unreachable: %v",
			liveNodes, unreachableNodes,
		)
	}

	leavingNodes := sets.NewString(ssInfo.LeavingNodes...)
	joiningNodes := sets.NewString(ssInfo.JoiningNodes...)
	movingNodes := sets.NewString(ssInfo.MovingNodes...)

	if setsIntersect(leavingNodes, joiningNodes, movingNodes) {
		return nil, fmt.Errorf(
			"unexpected state: some nodes were reported in more than one state. "+
				"Leaving: %v, "+
				"Joining: %v, "+
				"Moving: %v",
			leavingNodes, joiningNodes, movingNodes,
		)
	}

	for host, node := range nodes {
		switch {
		case liveNodes.Has(host):
			node.Status = NodeStatusUp
		case unreachableNodes.Has(host):
			node.Status = NodeStatusDown
		default:
			node.Status = NodeStatusUnknown
		}

		switch {
		case leavingNodes.Has(host):
			node.State = NodeStateLeaving
		case joiningNodes.Has(host):
			node.State = NodeStateJoining
		case movingNodes.Has(host):
			node.State = NodeStateMoving
		default:
			node.State = NodeStateNormal
		}

		if ssInfo.LocalHostId.String() == node.ID.String() {
			node.Local = true
		}
	}
	return nodes, nil
}

func (t *tool) Version() (*version.Version, error) {
	ssInfo, err := t.client.StorageService()
	if err != nil {
		return nil, err
	}
	return ssInfo.ReleaseVersion, nil
}
