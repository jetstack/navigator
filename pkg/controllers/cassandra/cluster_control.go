package cassandra

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
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
}

func NewControl(
	serviceControl service.Interface,
) ControlInterface {
	return &defaultCassandraClusterControl{
		serviceControl: serviceControl,
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	c = c.DeepCopy()
	glog.V(4).Infof("defaultCassandraClusterControl.Sync")
	if err := e.serviceControl.Sync(c); err != nil {
		glog.Errorf("error syncing service. %s", err)
		return err
	}
	return nil
}
