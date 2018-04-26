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

// NodeState represents the state of reachability of of a C* node from the
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

func (t *tool) Status() (NodeMap, error) {
	ssInfo, err := t.client.StorageService()
	if err != nil {
		return nil, err
	}

	nodes := NodeMap{}
	mappedNodes := sets.NewString()
	for host, id := range ssInfo.HostIdMap {
		mappedNodes.Insert(host)
		nodes[host] = &Node{
			Host:   host,
			ID:     id,
			Status: NodeStatusUnknown,
			State:  NodeStateNormal,
		}
	}

	liveNodes := sets.NewString(ssInfo.LiveNodes...)
	unreachableNodes := sets.NewString(ssInfo.UnreachableNodes...)
	if setsIntersect(liveNodes, unreachableNodes) {
		return nil, fmt.Errorf(
			"the sets of live and unreachable nodes should not intersect. "+
				"Live: %v, "+
				"Unreachable: %v",
			liveNodes, unreachableNodes,
		)
	}
	if !mappedNodes.IsSuperset(liveNodes.Union(unreachableNodes)) {
		return nil, fmt.Errorf(
			"mapped nodes must be a superset of Live and Unreachable nodes. "+
				"Live: %v, "+
				"Unreachable: %v, "+
				"Mapped: %v",
			liveNodes, unreachableNodes, mappedNodes,
		)
	}

	leavingNodes := sets.NewString(ssInfo.LeavingNodes...)
	joiningNodes := sets.NewString(ssInfo.JoiningNodes...)
	movingNodes := sets.NewString(ssInfo.MovingNodes...)

	if setsIntersect(leavingNodes, joiningNodes, movingNodes) {
		return nil, fmt.Errorf(
			"the sets of leaving, joining and moving nodes should not intersect. "+
				"Leaving: %v, "+
				"Joining: %v, "+
				"Moving: %v",
			leavingNodes, joiningNodes, movingNodes,
		)
	}

	if !mappedNodes.IsSuperset(leavingNodes.Union(joiningNodes).Union(movingNodes)) {
		return nil, fmt.Errorf(
			"mapped nodes must be a superset of leaving, joining and moving nodes. "+
				"Leaving: %v, "+
				"Joining: %v, "+
				"Moving: %v, "+
				"Mapped: %v",
			leavingNodes, joiningNodes, movingNodes, mappedNodes,
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
