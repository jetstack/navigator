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
	Plugins   []ElasticsearchClusterPlugin   `json:"plugins"`
	NodePools []ElasticsearchClusterNodePool `json:"nodePools"`
	Pilot     ElasticsearchPilotImage        `json:"pilot"`
	Image     ElasticsearchImage             `json:"image"`
	Sysctl    []string                       `json:"sysctl"`
}

// ElasticsearchClusterPlugin describes a specification of an ElasticsearchCluster plugin
// You must ensure the plugin is compatible with the version of Elasticsearch being deployed
// else the cluster will not deploy successfully
type ElasticsearchClusterPlugin struct {
	Name string `json:"name"`
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
	Phase         PilotPhase              `json:"phase"`
	Elasticsearch *PilotElasticsearchSpec `json:"elasticsearch"`
}

type PilotPhase string

const (
	PilotPhaseStarted        PilotPhase = "Started"
	PilotPhaseDecommissioned PilotPhase = "Decommissioned"
)

type PilotElasticsearchSpec struct {
	Plugins []ElasticsearchClusterPlugin `json:"plugins"`
	Roles   []ElasticsearchClusterRole   `json:"roles"`
}

type PilotStatus struct {
	Conditions []PilotCondition `json:"conditions"`
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
	// PilotConditionDecommissioned represents the fact that a given Pilot
	// condition is in a decommissioned state.
	PilotConditionDecommissioned PilotConditionType = "Decommissioned"
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
