package cassandra

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
	apiv1 "k8s.io/api/core/v1"
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
	serviceControl  service.Interface
	nodepoolControl nodepool.Interface
	recorder        record.EventRecorder
}

func NewControl(
	serviceControl service.Interface,
	nodepoolControl nodepool.Interface,
	recorder record.EventRecorder,
) ControlInterface {
	return &defaultCassandraClusterControl{
		serviceControl:  serviceControl,
		nodepoolControl: nodepoolControl,
		recorder:        recorder,
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	glog.V(4).Infof("defaultCassandraClusterControl.Sync")
	err := e.serviceControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncService,
			err,
		)
		return err
	}
	err = e.nodepoolControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			"cassandra.defaultCassandraClusterControl",
			ErrorSync,
			MessageErrorSyncNodePools,
			c,
		)
		return err
	}
	e.recorder.Event(
		c,
		apiv1.EventTypeNormal,
		SuccessSync,
		MessageSuccessSync,
	)
	return nil
}
