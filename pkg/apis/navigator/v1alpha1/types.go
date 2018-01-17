package v1alpha1

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

// ElasticsearchClusterStatus specifies the overall status of an
// ElasticsearchCluster.
type ElasticsearchClusterStatus struct {
	NodePools map[string]ElasticsearchClusterNodePoolStatus `json:"nodePools"`
	Health    ElasticsearchClusterHealth                    `json:"health"`
}

// ElasticsearchClusterNodePoolStatus specifies the status of a single node
// pool in an ElasticsearchCluster
type ElasticsearchClusterNodePoolStatus struct {
	// ReadyReplicas is the total number of ready pods in this cluster.
	ReadyReplicas int64 `json:"readyReplicas"`
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
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ElasticsearchCluster `json:"items"`
}

// ElasticsearchClusterSpec describes a specification for an ElasticsearchCluster
type ElasticsearchClusterSpec struct {
	// The version of Elasticsearch to be used for nodes in the cluster.
	Version semver.Version `json:"version"`

	// A list of plugins to install on nodes in the cluster.
	Plugins []string `json:"plugins"`

	// NodePools specify the various pools of nodes that make up this cluster.
	// There must be at least one master node specified.
	NodePools []ElasticsearchClusterNodePool `json:"nodePools"`

	// Pilot describes the image containing the pilot-elasticsearch binary to
	// run
	Pilot ElasticsearchPilotImage `json:"pilot"`

	// Image describes the Elasticsearch image to use
	Image *ElasticsearchImage `json:"image,omitempty"`

	// Sysctl can be used to specify a list of sysctl values to set on start-up
	// This can be used to set for example the vm.max_map_count parameter.
	Sysctl []string `json:"sysctl"`

	// The minimum number of masters required to form a quorum in the cluster.
	// If omitted, this will be set to a quorum of the master nodes in the
	// cluster. If set, the value *must* be greater than or equal to a quorum
	// of master nodes.
	MinimumMasters int64 `json:"minimumMasters,omitempty"`
}

// ElasticsearchClusterNodePool describes a node pool within an ElasticsearchCluster.
// The nodes in this pool will be configured to be of the specified roles
type ElasticsearchClusterNodePool struct {
	// Name of the node pool.
	Name string `json:"name"`

	// Number of replicas in the pool.
	Replicas int64 `json:"replicas"`

	// Roles that nodes in this pool should perform within the cluster.
	Roles []ElasticsearchClusterRole `json:"roles"`

	// NodeSelector should be specified to force nodes in this pool to run on
	// nodes matching the given selector.
	NodeSelector map[string]string `json:"nodeSelector"`

	// Resources specifies the resource requirements to be used for nodes that
	// are part of the pool.
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence specifies the configuration for persistent data for this
	// node. Disabling persistence can cause issues when nodes restart, so
	// should only be using for testing purposes.
	Persistence ElasticsearchClusterPersistenceConfig `json:"persistence,omitempty"`
}

// ElasticsearchClusterRole is a node role in an ElasticsearchCluster.
type ElasticsearchClusterRole string

const (
	ElasticsearchRoleData   ElasticsearchClusterRole = "data"
	ElasticsearchRoleMaster ElasticsearchClusterRole = "master"
	ElasticsearchRoleIngest ElasticsearchClusterRole = "ingest"
)

// ElasticsearchClusterPersistenceConfig contains persistent volume
// configuration.
type ElasticsearchClusterPersistenceConfig struct {
	// Toggle whether persistence should be enabled on this cluster. If false,
	// an emptyDir will be provisioned to store Elasticsearch data.
	Enabled bool `json:"enabled"`

	// Size of the persistent volume to provision (required if persistence is
	// enabled).
	Size resource.Quantity `json:"size"`

	// StorageClass to use for the persistent volume claim. If not set, the
	// default cluster storage class will be used.
	StorageClass string `json:"storageClass"`
}

// ImageSpec specifies a docker image to be used.
type ImageSpec struct {
	// Repository is the repository of a docker image (e.g. 'alpine').
	Repository string `json:"repository"`

	// Tag is the tag of a docker image (e.g. 'latest').
	Tag string `json:"tag"`

	// PullPolicy is the policy for pulling docker images. If not set, the
	// cluster default will be used.
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

type ElasticsearchPilotImage struct {
	ImageSpec `json:",inline"`
}

type ElasticsearchImage struct {
	ImageSpec `json:",inline"`
	// FsGroup specifies the user that the should be set for the pods fsGroup
	FsGroup int64 `json:"fsGroup"`
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
