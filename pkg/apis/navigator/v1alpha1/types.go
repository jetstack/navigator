package v1alpha1

import (
	"github.com/coreos/go-semver/semver"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/api/version"
)

const (
	ElasticsearchClusterNameLabel          = "navigator.jetstack.io/elasticsearch-cluster-name"
	ElasticsearchNodePoolNameLabel         = "navigator.jetstack.io/elasticsearch-node-pool-name"
	ElasticsearchNodePoolVersionAnnotation = "navigator.jetstack.io/elasticsearch-version"
	ElasticsearchRoleLabelPrefix           = "navigator.jetstack.io/elasticsearch-role-"

	CassandraClusterNameLabel  = "navigator.jetstack.io/cassandra-cluster-name"
	CassandraNodePoolNameLabel = "navigator.jetstack.io/cassandra-node-pool-name"
	PilotLabel                 = "navigator.jetstack.io/has-pilot"
)

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
	NavigatorClusterConfig `json:",inline"`

	// List of node pools belonging to this CassandraCluster.
	// NodePools cannot currently be removed.
	// A cluster with 0 node pools will not function correctly.
	// +optional
	NodePools []CassandraClusterNodePool `json:"nodePools,omitempty"`

	// Image describes the Cassandra database image to use.
	// This should only be set if version auto-detection is not desired.
	// If set, the image tag used must correspond to the version specified
	// in 'spec.version'.
	// This is an advanced feature and should be used with caution. It is
	// the end user's repsonsibility to ensure the 'Image' here matches with
	// the Version specified below.
	// +optional
	Image *ImageSpec `json:"image,omitempty"`

	// The version of the database to be used for nodes in the cluster.
	// This field must be a valid Cassandra version, e.g. '3.11.2'.
	Version version.Version `json:"version"`
}

// CassandraClusterNodePool describes a node pool within a CassandraCluster.
type CassandraClusterNodePool struct {
	// Name of the node pool specified as a DNS_LABEL.
	// Each node pool must specify a name.
	Name string `json:"name"`

	// The number of desired replicas in this node pool.
	// This value will correspond to the number of replicas in the created
	// StatefulSet.
	// If not set, a default of 1 will be used.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Persistence specifies the persistent volume configuration for this node.
	// Disabling persistence can cause issues when a node restarts.
	// Cannot be updated.
	// +optional
	Persistence *PersistenceConfig `json:"persistence,omitempty"`

	// NodeSelector should be specified to force nodes in this pool to run on
	// nodes matching the given selector.
	// In future this may be superceded by an 'affinity' field.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Rack specifies the cassandra rack with which to label nodes in this
	// nodepool.
	// If this is not set, the name of the node pool will be used.
	// +optional
	Rack *string `json:"rack,omitempty"`

	// Datacenter specifies the cassandra datacenter with which to label nodes
	// in this nodepool. If this is not set, a default will be selected.
	// If this is not set, the default of 'navigator-default-datacenter' will
	// be used.
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`

	// Resources specifies the resource requirements to be used for nodes that
	// are part of the pool.
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`
}

type CassandraClusterStatus struct {
	NavigatorClusterStatus `json:",inline"`

	// The status of each node pool in the CassandraCluster.
	// This will be periodically updated by Navigator.
	NodePools map[string]CassandraClusterNodePoolStatus `json:"nodePools,omitempty"`
}

type CassandraClusterNodePoolStatus struct {
	// The number of replicas in the node pool that are currently 'Ready'.
	ReadyReplicas int32 `json:"readyReplicas"`
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
	NavigatorClusterStatus `json:",inline"`

	// The status of each node pool in the ElasticsearchCluster.
	// This will be periodically updated by Navigator.
	NodePools map[string]ElasticsearchClusterNodePoolStatus `json:"nodePools,omitempty"`
	// The health of the ElasticsearchCluster.
	// This will be one of Red, Yellow or Green.
	// +optional
	Health *ElasticsearchClusterHealth `json:"health,omitempty"`
}

// ElasticsearchClusterNodePoolStatus specifies the status of a single node
// pool in an ElasticsearchCluster
type ElasticsearchClusterNodePoolStatus struct {
	// ReadyReplicas is the total number of ready pods in this cluster.
	ReadyReplicas int32 `json:"readyReplicas"`
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
	NavigatorClusterConfig `json:",inline"`

	// The version of the database to be used for nodes in the cluster.
	// This field must be a valid Elasticsearch version, e.g. '6.1.1'.
	Version semver.Version `json:"version"`

	// List of plugins to install on nodes in the cluster.
	// pilot-elasticsearch will install these plugins when each node is started.
	// +optional
	Plugins []string `json:"plugins,omitempty"`

	// List of node pools belonging to this ElasticsearchCluster.
	// NodePools cannot currently be removed.
	// There must be at least one master node to form a valid cluster.
	// +optional
	NodePools []ElasticsearchClusterNodePool `json:"nodePools,omitempty"`

	// Image describes the Elasticsearch database image to use.
	// This should only be set if version auto-detection is not desired.
	// If set, the image tag used must correspond to the version specified
	// in 'spec.version'.
	// +optional
	Image *ImageSpec `json:"image,omitempty"`

	// The minimum number of masters required to form a quorum in the cluster.
	// If omitted, this will be set to a quorum of the master nodes in the
	// cluster.
	// If set, the value *must* be greater than or equal to a quorum of master
	// nodes.
	// +optional
	MinimumMasters *int32 `json:"minimumMasters,omitempty"`
}

// ElasticsearchClusterNodePool describes a node pool within an ElasticsearchCluster.
// The nodes in this pool will be configured to be of the specified roles
type ElasticsearchClusterNodePool struct {
	// Name of the node pool specified as a DNS_LABEL.
	// Each node pool must specify a name.
	Name string `json:"name"`

	// The number of desired replicas in this node pool.
	// This value will correspond to the number of replicas in the created
	// StatefulSet.
	// If not set, a default of 1 will be used.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Roles that nodes in this pool should perform within the cluster.
	// At least one role must be specified.
	Roles []ElasticsearchClusterRole `json:"roles,omitempty"`

	// NodeSelector should be specified to force nodes in this pool to run on
	// nodes matching the given selector.
	// In future this may be superceded by an 'affinity' field.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Resources specifies the resource requirements to be used for nodes that
	// are part of the pool.
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence specifies the persistent volume configuration for this node.
	// Disabling persistence can cause issues when a node restarts.
	// Cannot be updated.
	// +optional
	Persistence *PersistenceConfig `json:"persistence,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`
}

// ElasticsearchClusterRole is a node role in an ElasticsearchCluster.
type ElasticsearchClusterRole string

const (
	ElasticsearchRoleData   ElasticsearchClusterRole = "data"
	ElasticsearchRoleMaster ElasticsearchClusterRole = "master"
	ElasticsearchRoleIngest ElasticsearchClusterRole = "ingest"
)

// PersistenceConfig contains persistent volume configuration.
type PersistenceConfig struct {
	// Size of the persistent volume to provision (required if persistence is
	// enabled).
	Size resource.Quantity `json:"size"`

	// StorageClass to use for the persistent volume claim. If not set, the
	// default cluster storage class will be used.
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`
}

// ImageSpec specifies a docker image to be used.
type ImageSpec struct {
	// Repository is the repository of a docker image (e.g. 'alpine').
	Repository string `json:"repository"`

	// Tag is the tag of a docker image (e.g. 'latest').
	Tag string `json:"tag"`

	// PullPolicy is the policy for pulling docker images. If not set, the
	// cluster default will be used.
	// +optional
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`
}

type NavigatorClusterConfig struct {
	// Pilot describes the pilot image to use.
	// This field is currently required.
	PilotImage ImageSpec `json:"pilotImage"`

	// Security related options to be applied to the cluster pods.
	SecurityContext NavigatorSecurityContext `json:"securityContext"`

	// If set to true, no actions will take place on this cluster.
	Paused bool `json:"paused,omitempty"`
}

type NavigatorClusterStatus struct {
	// Represents the latest available observations of a cluster's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ClusterCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type NavigatorSecurityContext struct {
	// Optional user to run the pilot process as.
	// This will also be used as the FSGroup parameter for created pods.
	// Should correspond to the value set as part of the Dockerfile when
	// a manual image override has been specified.
	// +optional
	RunAsUser *int64 `json:"runAsUser,omitempty"`
}

type ClusterConditionType string

const (
	ClusterConditionAvailable ClusterConditionType = "Available"

	ClusterConditionProgressing ClusterConditionType = "Progressing"
)

const (
	// PausedClusterReason is added in a cluster when it is paused. Lack of progress shouldn't be
	// estimated once a cluster is paused.
	PausedClusterReason = "ClusterPaused"
	// ResumedClusterReason is added in a cluster when it is resumed.
	ResumedClusterReason = "ClusterResumed"
)

type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
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
}

type PilotStatus struct {
	// Conditions representing the current state of the Pilot.
	// +optional
	Conditions []PilotCondition `json:"conditions,omitempty"`

	// Status information specific to Elasticsearch Pilots.
	// +optional
	Elasticsearch *ElasticsearchPilotStatus `json:"elasticsearch,omitempty"`

	// Status information specific to Cassandra Pilots.
	// +optional
	Cassandra *CassandraPilotStatus `json:"cassandra,omitempty"`
}

type ElasticsearchPilotStatus struct {
	// Documents is the current number of documents on this node.
	// A nil value indicates an unknown number of documents.
	// Zero documents indicates the node is empty.
	// +optional
	Documents *int64 `json:"documents,omitempty"`

	// Version as reported by the Elasticsearch process
	// This field may be nil if the version number is not currently known.
	// +optional
	Version *semver.Version `json:"version,omitempty"`
}

type CassandraPilotStatus struct {
	// Version as reported by the Elasticsearch process
	// This field may be nil if the version number is not currently known.
	// +optional
	Version *version.Version `json:"version,omitempty"`
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
