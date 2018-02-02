package navigator

import (
	"github.com/coreos/go-semver/semver"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// In this file we define the outer containing types for the ElasticsearchCluster
// type. We could import these directly into message types defined in the types.proto
// file, but this is still TODO

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraCluster describes a specification for an Cassandra cluster
type CassandraCluster struct {
	// we embed these types so the CassandraCluster implements runtime.Object
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   CassandraClusterSpec
	Status CassandraClusterStatus
}

type CassandraClusterSpec struct {
	NavigatorClusterConfig

	NodePools []CassandraClusterNodePool
	Image     ImageSpec
	CqlPort   int32
}

type CassandraClusterNodePool struct {
	Name     string
	Replicas int64
}

type CassandraClusterStatus struct {
	NodePools map[string]CassandraClusterNodePoolStatus
}

type CassandraClusterNodePoolStatus struct {
	ReadyReplicas int64
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraClusterList defines a List type for our custom CassandraCluster type.
// This is needed in order to make List operations work.
type CassandraClusterList struct {
	// we embed these types so that CassandraClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta
	metav1.ListMeta

	Items []CassandraCluster
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchCluster describes a specification for an Elasticsearch cluster
type ElasticsearchCluster struct {
	// we embed these types so the ElasticsearchCluster implements runtime.Object
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ElasticsearchClusterSpec
	Status ElasticsearchClusterStatus
}

type ElasticsearchClusterStatus struct {
	NodePools map[string]ElasticsearchClusterNodePoolStatus
	Health    ElasticsearchClusterHealth
}

type ElasticsearchClusterNodePoolStatus struct {
	ReadyReplicas int32
}

type ElasticsearchClusterHealth string

const (
	ElasticsearchClusterHealthRed    ElasticsearchClusterHealth = "Red"
	ElasticsearchClusterHealthYellow ElasticsearchClusterHealth = "Yellow"
	ElasticsearchClusterHealthGreen  ElasticsearchClusterHealth = "Green"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchClusterList defines a List type for our custom ElasticsearchCluster type.
// This is needed in order to make List operations work.
type ElasticsearchClusterList struct {
	// we embed these types so that ElasticsearchClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta
	metav1.ListMeta

	Items []ElasticsearchCluster
}

type ElasticsearchClusterSpec struct {
	NavigatorClusterConfig

	Version        semver.Version
	Image          *ImageSpec
	Plugins        []string
	NodePools      []ElasticsearchClusterNodePool
	MinimumMasters int32
}

type ElasticsearchClusterNodePool struct {
	Name         string
	Replicas     int32
	Roles        []ElasticsearchClusterRole
	NodeSelector map[string]string
	Resources    v1.ResourceRequirements
	Persistence  ElasticsearchClusterPersistenceConfig
}

type ElasticsearchClusterRole string

const (
	ElasticsearchRoleData   ElasticsearchClusterRole = "data"
	ElasticsearchRoleMaster ElasticsearchClusterRole = "master"
	ElasticsearchRoleIngest ElasticsearchClusterRole = "ingest"
)

type ElasticsearchClusterPersistenceConfig struct {
	Enabled      bool
	Size         resource.Quantity
	StorageClass string
}

type ImageSpec struct {
	Repository string
	Tag        string
	PullPolicy v1.PullPolicy
}

type NavigatorClusterConfig struct {
	PilotImage ImageSpec

	SecurityContext NavigatorSecurityContext

	Sysctls []string
}

type NavigatorSecurityContext struct {
	RunAsUser *int64
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Pilot struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   PilotSpec
	Status PilotStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PilotList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Pilot
}

type PilotSpec struct {
	Elasticsearch *PilotElasticsearchSpec
}

type PilotPhase string

const (
	PilotPhasePreStart  PilotPhase = "PreStart"
	PilotPhasePostStart PilotPhase = "PostStart"
	PilotPhasePreStop   PilotPhase = "PreStop"
	PilotPhasePostStop  PilotPhase = "PostStop"
)

type PilotElasticsearchSpec struct {
}

type PilotStatus struct {
	LastCompletedPhase PilotPhase
	Conditions         []PilotCondition
	// Contains status information specific to Elasticsearch Pilots
	Elasticsearch *ElasticsearchPilotStatus
}

type ElasticsearchPilotStatus struct {
	// Documents is the current number of documents on this node. nil indicates
	// an unknown number of documents, whereas 0 indicates that the node is
	// empty
	Documents *int64
	Version   string
}

// PilotCondition contains condition information for a Pilot.
type PilotCondition struct {
	// Type of the condition, currently ('Ready').
	Type PilotConditionType

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// PilotConditionType represents a Pilot condition value.
type PilotConditionType string

const (
	// PilotConditionReady represents the fact that a given Pilot condition
	// is in ready state.
	PilotConditionReady PilotConditionType = "Ready"
	// PilotConditionStarted represents the fact that a given Pilot condition
	// is in started state.
	PilotConditionStarted PilotConditionType = "Started"
	// PilotConditionStopped represents the fact that a given Pilot
	// condition is in a stopped state.
	PilotConditionStopped PilotConditionType = "Stopped"
)

// ConditionStatus represents a condition's status.
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)
