package cassandra

import (
	"github.com/golang/glog"
	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	servicecql "github.com/jetstack/navigator/pkg/controllers/cassandra/service/cql"
	serviceseedprovider "github.com/jetstack/navigator/pkg/controllers/cassandra/service/seedprovider"
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
	MessageErrorSyncPilots         = "Error syncing pilots: %s"
	MessageSuccessSync             = "Successfully synced CassandraCluster"
)

type ControlInterface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

var _ ControlInterface = &defaultCassandraClusterControl{}

type defaultCassandraClusterControl struct {
	seedProviderServiceControl serviceseedprovider.Interface
	cqlServiceControl          servicecql.Interface
	nodepoolControl            nodepool.Interface
	pilotControl               pilot.Interface
	recorder                   record.EventRecorder
}

func NewControl(
	seedProviderServiceControl serviceseedprovider.Interface,
	cqlServiceControl servicecql.Interface,
	nodepoolControl nodepool.Interface,
	pilotControl pilot.Interface,
	recorder record.EventRecorder,
) ControlInterface {
	return &defaultCassandraClusterControl{
		seedProviderServiceControl: seedProviderServiceControl,
		cqlServiceControl:          cqlServiceControl,
		nodepoolControl:            nodepoolControl,
		pilotControl:               pilotControl,
		recorder:                   recorder,
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	err := e.seedProviderServiceControl.Sync(c)
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
	glog.V(4).Infof("Synced seed service")

	err = e.cqlServiceControl.Sync(c)
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
	glog.V(4).Infof("Synced CQL service")

	err = e.nodepoolControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncNodePools,
			err,
		)
		return err
	}
	glog.V(4).Infof("Synced statefulsets")

	err = e.pilotControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncPilots,
			err,
		)
		return err
	}
	glog.V(4).Infof("Synced pilots")

	e.recorder.Event(
		c,
		apiv1.EventTypeNormal,
		SuccessSync,
		MessageSuccessSync,
	)
	return nil
}
