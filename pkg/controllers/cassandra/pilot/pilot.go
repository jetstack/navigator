package pilot

import (
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/api/version"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type pilotControl struct {
	naviClient navigator.Interface
	pilots     navlisters.PilotLister
	recorder   record.EventRecorder
}

var _ Interface = &pilotControl{}

func NewControl(
	naviClient navigator.Interface,
	pilots navlisters.PilotLister,
	recorder record.EventRecorder,
) *pilotControl {
	return &pilotControl{
		naviClient: naviClient,
		pilots:     pilots,
		recorder:   recorder,
	}

}

// Sync
func (c *pilotControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	return c.updateDiscoveredVersions(cluster)
}

func (c *pilotControl) updateDiscoveredVersions(cluster *v1alpha1.CassandraCluster) error {
	glog.V(4).Infof("updateDiscoveredVersions for cluster: %s/%s", cluster.Namespace, cluster.Name)
	selector, err := util.SelectorForCluster(cluster)
	if err != nil {
		return err
	}
	pilots, err := c.pilots.List(selector)
	if err != nil {
		return err
	}
	if len(pilots) < 1 {
		glog.V(4).Infof("No pilots found matching selector: %s", selector)
	}
	for _, pilot := range pilots {
		nodePoolNameForPilot, nodePoolNameFound := pilot.Labels[v1alpha1.CassandraNodePoolNameLabel]
		if !nodePoolNameFound {
			glog.Warningf("Skipping pilot without NodePoolNameLabelKey: %s", pilot.Name)
			continue
		}
		nodePoolStatus := cluster.Status.NodePools[nodePoolNameForPilot]
		switch {
		case pilot.Status.Cassandra == nil:
			glog.V(4).Infof(
				"Pilot %s/%s has no status. Setting nodepool version to nil",
				pilot.Namespace, pilot.Name,
			)
			nodePoolStatus.Version = nil
		case pilot.Status.Cassandra.Version == nil:
			glog.V(4).Infof(
				"Pilot %s/%s has not reported its version. Setting nodepool version to nil",
				pilot.Namespace, pilot.Name,
			)
			nodePoolStatus.Version = nil
		case nodePoolStatus.Version == nil:
			nodePoolStatus.Version = pilot.Status.Cassandra.Version
		case pilot.Status.Cassandra.Version.LessThan(nodePoolStatus.Version):
			glog.V(4).Infof(
				"Found lower pilot version: %s, %s",
				nodePoolNameForPilot, pilot.Status.Cassandra.Version,
			)
			nodePoolStatus.Version = pilot.Status.Cassandra.Version
		}
		cluster.Status.NodePools[nodePoolNameForPilot] = nodePoolStatus
	}
	return nil
}

func UpdateLabels(
	o metav1.Object,
	newLabels map[string]string,
) {
	labels := o.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for key, val := range newLabels {
		labels[key] = val
	}
	o.SetLabels(labels)
}

type PilotBuilder struct {
	pilot *v1alpha1.Pilot
}

func NewPilotBuilder() *PilotBuilder {
	return &PilotBuilder{
		pilot: &v1alpha1.Pilot{},
	}
}

func (pb *PilotBuilder) ForCluster(cluster metav1.Object) *PilotBuilder {
	UpdateLabels(pb.pilot, util.ClusterLabels(cluster))
	pb.pilot.SetNamespace(cluster.GetNamespace())
	pb.pilot.SetOwnerReferences(
		append(
			pb.pilot.GetOwnerReferences(),
			util.NewControllerRef(cluster),
		),
	)
	return pb
}

func (pb *PilotBuilder) ForNodePool(np *v1alpha1.CassandraClusterNodePool) *PilotBuilder {
	UpdateLabels(
		pb.pilot,
		map[string]string{
			v1alpha1.CassandraNodePoolNameLabel: np.Name,
		},
	)
	return pb
}

func (pb *PilotBuilder) WithCassandraStatus() *PilotBuilder {
	pb.pilot.Status.Cassandra = &v1alpha1.CassandraPilotStatus{}
	return pb
}

func (pb *PilotBuilder) WithDiscoveredCassandraVersion(v string) *PilotBuilder {
	pb.WithCassandraStatus()
	pb.pilot.Status.Cassandra.Version = version.New(v)
	return pb
}

func (pb *PilotBuilder) Build() *v1alpha1.Pilot {
	return pb.pilot
}
