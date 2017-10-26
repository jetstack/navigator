package cassandra

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"k8s.io/client-go/tools/record"
)

const (
	errorSync = "ErrSync"

	successSync = "SuccessSync"

	messageErrorSyncServiceAccount = "Error syncing service account: %s"
	messageErrorSyncConfigMap      = "Error syncing config map: %s"
	messageErrorSyncService        = "Error syncing service: %s"
	messageErrorSyncNodePools      = "Error syncing node pools: %s"
	messageSuccessSync             = "Successfully synced CassandraCluster"
)

type ControlInterface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

var _ ControlInterface = &defaultCassandraClusterControl{}

type defaultCassandraClusterControl struct {
	recorder record.EventRecorder
}

func NewControl(recorder record.EventRecorder) ControlInterface {
	return &defaultCassandraClusterControl{
		recorder: recorder,
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	glog.V(4).Infof("defaultCassandraClusterControl.Sync")
	e.recorder.Event(
		c,
		"cassandra.defaultCassandraClusterControl",
		successSync,
		messageSuccessSync,
	)
	return nil
}
