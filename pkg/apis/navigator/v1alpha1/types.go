package v1alpha1

import (
	"k8s.io/api/core/v1"
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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CassandraClusterSpec   `json:"spec"`
	Status CassandraClusterStatus `json:"status"`
}

type CassandraClusterSpec struct {
	Sysctl     []string                   `json:"sysctl"`
	NodePools  []CassandraClusterNodePool `json:"nodePools"`
	Image      ImageSpec                  `json:"image"`
	PilotImage ImageSpec                  `json:"pilotImage"`
	CqlPort    int32                      `json:"cqlPort"`
}

// CassandraClusterNodePool describes a node pool within a CassandraCluster.
type CassandraClusterNodePool struct {
	Name     string `json:"name"`
	Replicas int64  `json:"replicas"`
}

type CassandraClusterStatus struct {
	NodePools map[string]CassandraClusterNodePoolStatus `json:"nodePools"`
}

type CassandraClusterNodePoolStatus struct {
	ReadyReplicas int64 `json:"readyReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraClusterList defines a List type for our custom CassandraCluster type.
// This is needed in order to make List operations work.
type CassandraClusterList struct {
	// we embed these types so that CassandraClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CassandraCluster `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchCluster describes a specification for an Elasticsearch cluster
type ElasticsearchCluster struct {
	// we embed these types so the ElasticsearchCluster implements runtime.Object
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ElasticsearchClusterSpec   `json:"spec"`
	Status ElasticsearchClusterStatus `json:"status"`
}

type ElasticsearchClusterStatus struct {
	NodePools map[string]ElasticsearchClusterNodePoolStatus `json:"nodePools"`
}

type ElasticsearchClusterNodePoolStatus struct {
	ReadyReplicas int64 `json:"readyReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchClusterList defines a List type for our custom ElasticsearchCluster type.
// This is needed in order to make List operations work.
type ElasticsearchClusterList struct {
	// we embed these types so that ElasticsearchClusterList implements runtime.Object and List interfaces
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ElasticsearchCluster `json:"items"`
}

// ElasticsearchClusterSpec describes a specification for an ElasticsearchCluster
type ElasticsearchClusterSpec struct {
	Plugins   []string                       `json:"plugins"`
	NodePools []ElasticsearchClusterNodePool `json:"nodePools"`
	Pilot     ElasticsearchPilotImage        `json:"pilot"`
	Image     ElasticsearchImage             `json:"image"`
	Sysctl    []string                       `json:"sysctl"`
}

// ElasticsearchClusterNodePool describes a node pool within an ElasticsearchCluster.
// The nodes in this pool will be configured to be of the specified roles
type ElasticsearchClusterNodePool struct {
	Name         string                                `json:"name"`
	Replicas     int64                                 `json:"replicas"`
	Roles        []ElasticsearchClusterRole            `json:"roles"`
	NodeSelector map[string]string                     `json:"nodeSelector"`
	Resources    *v1.ResourceRequirements              `json:"resources,omitempty"`
	Persistence  ElasticsearchClusterPersistenceConfig `json:"persistence,omitempty"`
	// Config is a map of configuration files to be placed in the elasticsearch
	// config directory. Environment variables may be used in these files and
	// they will be automatically expanded by the Elasticsearch process.
	Config map[string]string `json:"config"`
}

type ElasticsearchClusterRole string

const (
	ElasticsearchRoleData   ElasticsearchClusterRole = "data"
	ElasticsearchRoleMaster ElasticsearchClusterRole = "master"
	ElasticsearchRoleIngest ElasticsearchClusterRole = "ingest"
)

type ElasticsearchClusterPersistenceConfig struct {
	Enabled      bool   `json:"enabled"`
	Size         string `json:"size"`
	StorageClass string `json:"storageClass"`
}

type ImageSpec struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
}

type ElasticsearchPilotImage struct {
	ImageSpec `json:",inline"`
}

type ElasticsearchImage struct {
	ImageSpec `json:",inline"`
	FsGroup   int64 `json:"fsGroup"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Pilot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PilotSpec   `json:"spec"`
	Status PilotStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PilotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Pilot `json:"items"`
}

type PilotSpec struct {
	Elasticsearch *PilotElasticsearchSpec `json:"elasticsearch"`
}

type PilotPhase string

const (
	// PreStart occurs before the Pilot's subprocess has been started.
	PilotPhasePreStart PilotPhase = "PreStart"
	// PostStart occurs immediately after the Pilot's subprocess has been
	// started.
	PilotPhasePostStart PilotPhase = "PostStart"
	// PreStop occurs just before the Pilot's subprocess is sent a graceful
	// termination signal. These hooks will block termination of the process.
	PilotPhasePreStop PilotPhase = "PreStop"
	// PostStop occurs after the Pilot's has stopped. These can be used to
	// clean up, or whatever other action that may need to be performed.
	PilotPhasePostStop PilotPhase = "PostStop"
)

type PilotElasticsearchSpec struct {
}

type PilotStatus struct {
	LastCompletedPhase PilotPhase       `json:"lastCompletedPhase"`
	Conditions         []PilotCondition `json:"conditions"`
	// Contains status information specific to Elasticsearch Pilots
	Elasticsearch *ElasticsearchPilotStatus `json:"elasticsearch,omitempty"`
}

type ElasticsearchPilotStatus struct {
	// Documents is the current number of documents on this node. nil indicates
	// an unknown number of documents, whereas 0 indicates that the node is
	// empty
	Documents *int64 `json:"documents,omitempty"`
}

// PilotCondition contains condition information for a Pilot.
type PilotCondition struct {
	// Type of the condition, currently ('Ready').
	Type PilotConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string `json:"reason"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string `json:"message"`
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
