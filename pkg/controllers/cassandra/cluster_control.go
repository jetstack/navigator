package cassandra

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/controllers/cassandra/service"
	"k8s.io/client-go/tools/record"
)

const (
	ErrorSync = "ErrSync"

	SuccessSync = "SuccessSync"

	MessageErrorSyncServiceAccount = "Error syncing service account: %s"
	MessageErrorSyncConfigMap      = "Error syncing config map: %s"
	MessageErrorSyncService        = "Error syncing service: %s"
	MessageErrorSyncNodePools      = "Error syncing node pools: %s"
	MessageSuccessSync             = "Successfully synced CassandraCluster"
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
			ErrorSync,
			MessageErrorSyncService,
			c,
		)
		return err
	}
	e.recorder.Event(
		c,
		"cassandra.defaultCassandraClusterControl",
		SuccessSync,
		MessageSuccessSync,
	)
	return nil
}
