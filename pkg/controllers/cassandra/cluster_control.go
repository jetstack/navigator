package cassandra

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
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
	serviceControl service.Interface
	recorder       record.EventRecorder
}

func NewControl(serviceControl service.Interface, recorder record.EventRecorder) ControlInterface {
	return &defaultCassandraClusterControl{
		serviceControl: serviceControl,
		recorder:       recorder,
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	glog.V(4).Infof("defaultCassandraClusterControl.Sync")
	err := e.serviceControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			"cassandra.defaultCassandraClusterControl",
			errorSync,
			messageErrorSyncService,
			c,
		)
		return err
	}
	e.recorder.Event(
		c,
		"cassandra.defaultCassandraClusterControl",
		successSync,
		messageSuccessSync,
	)
	return nil
}
