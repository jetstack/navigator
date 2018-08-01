package cassandra

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/role"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/rolebinding"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/seedlabeller"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/service"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/serviceaccount"
)

const (
	ErrorSync   = "ErrSync"
	SuccessSync = "SuccessSync"

	MessageErrorSyncServiceAccount = "Error syncing service account: %s"
	MessageErrorSyncRole           = "Error syncing role: %s"
	MessageErrorSyncRoleBinding    = "Error syncing role binding: %s"
	MessageErrorSyncConfigMap      = "Error syncing config map: %s"
	MessageErrorSyncService        = "Error syncing service: %s"
	MessageErrorSyncNodePools      = "Error syncing node pools: %s"
	MessageErrorSyncSeedLabels     = "Error syncing seed labels: %s"
	MessageErrorSyncPilots         = "Error syncing pilots: %s"
	MessageErrorSync               = "Error syncing: %s"
	MessageSuccessSync             = "Successfully synced CassandraCluster"
)

type ControlInterface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

var _ ControlInterface = &defaultCassandraClusterControl{}

type defaultCassandraClusterControl struct {
	seedProviderServiceControl service.Interface
	nodesServiceControl        service.Interface
	nodepoolControl            nodepool.Interface
	serviceAccountControl      serviceaccount.Interface
	roleControl                role.Interface
	roleBindingControl         rolebinding.Interface
	seedLabellerControl        seedlabeller.Interface
	pilotControl               pilot.Interface
	recorder                   record.EventRecorder
	state                      *controllers.State
}

func NewControl(
	seedProviderServiceControl service.Interface,
	nodesServiceControl service.Interface,
	nodepoolControl nodepool.Interface,
	serviceAccountControl serviceaccount.Interface,
	roleControl role.Interface,
	roleBindingControl rolebinding.Interface,
	seedlabellerControl seedlabeller.Interface,
	pilotControl pilot.Interface,
	recorder record.EventRecorder,
	state *controllers.State,
) ControlInterface {
	return &defaultCassandraClusterControl{
		seedProviderServiceControl: seedProviderServiceControl,
		nodesServiceControl:        nodesServiceControl,
		nodepoolControl:            nodepoolControl,
		serviceAccountControl:      serviceAccountControl,
		roleControl:                roleControl,
		roleBindingControl:         roleBindingControl,
		seedLabellerControl:        seedlabellerControl,
		pilotControl:               pilotControl,
		recorder:                   recorder,
		state:                      state,
	}
}

// syncPausedConditions checks if the given cluster is paused or not and adds an appropriate condition.
func (e *defaultCassandraClusterControl) syncPausedConditions(c *v1alpha1.CassandraCluster) {
	cond := c.Status.GetStatusCondition(v1alpha1.ClusterConditionProgressing)
	pausedCondExists := cond != nil && cond.Reason == v1alpha1.PausedClusterReason

	if c.Spec.Paused && !pausedCondExists {
		c.Status.UpdateStatusCondition(
			v1alpha1.ClusterConditionProgressing,
			v1alpha1.ConditionFalse,
			v1alpha1.PausedClusterReason,
			"Cluster is paused",
		)
	} else if !c.Spec.Paused && pausedCondExists {
		c.Status.UpdateStatusCondition(
			v1alpha1.ClusterConditionProgressing,
			v1alpha1.ConditionTrue,
			v1alpha1.ResumedClusterReason,
			"Cluster is resumed",
		)
	}
}

func (e *defaultCassandraClusterControl) Sync(c *v1alpha1.CassandraCluster) error {
	var err error

	e.syncPausedConditions(c)

	if c.Spec.Paused == true {
		glog.Infof("defaultCassandraClusterControl.Sync skipped, since cluster is paused")
		return nil
	}

	glog.V(4).Infof("defaultCassandraClusterControl.Sync")
	err = e.seedProviderServiceControl.Sync(c)
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
	err = e.nodesServiceControl.Sync(c)
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
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncNodePools,
			err,
		)
		return err
	}
	err = e.serviceAccountControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncServiceAccount,
			err,
		)
		return err
	}
	err = e.roleControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncRole,
			err,
		)
		return err
	}
	err = e.roleBindingControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncRoleBinding,
			err,
		)
		return err
	}
	err = e.seedLabellerControl.Sync(c)
	if err != nil {
		e.recorder.Eventf(
			c,
			apiv1.EventTypeWarning,
			ErrorSync,
			MessageErrorSyncSeedLabels,
			err,
		)
		return err
	}

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

	a := NextAction(c)
	if a != nil {
		glog.V(4).Infof("Executing action: %#v", a)
		err = a.Execute(e.state)
		if err != nil {
			e.recorder.Eventf(
				c,
				apiv1.EventTypeWarning,
				ErrorSync,
				MessageErrorSync,
				err,
			)
			return errors.Wrap(err, "failure while executing action")
		}
	}
	e.recorder.Event(
		c,
		apiv1.EventTypeNormal,
		SuccessSync,
		MessageSuccessSync,
	)
	return nil
}

func NextAction(c *v1alpha1.CassandraCluster) controllers.Action {
	for _, np := range c.Spec.NodePools {
		_, found := c.Status.NodePools[np.Name]
		if !found {
			return &actions.CreateNodePool{
				Cluster:  c,
				NodePool: &np,
			}
		}
	}
	for _, np := range c.Spec.NodePools {
		nps := c.Status.NodePools[np.Name]
		if *np.Replicas > nps.ReadyReplicas {
			return &actions.ScaleOut{
				Cluster:  c,
				NodePool: &np,
			}
		}
	}
	for _, np := range c.Spec.NodePools {
		nps := c.Status.NodePools[np.Name]
		if nps.Version == nil {
			return nil
		}
		if c.Spec.Version.LessThan(nps.Version) {
			glog.Error("Version downgrades are not supported")
			return nil
		}
		if nps.Version.Major() != c.Spec.Version.Major() {
			glog.Error("Major version upgrades are not supported")
			return nil
		}
		if nps.Version.LessThan(&c.Spec.Version) {
			return &actions.UpdateVersion{
				Cluster:  c,
				NodePool: &np,
			}
		}
	}
	return nil
}
