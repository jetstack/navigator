package couchbase

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"k8s.io/client-go/tools/record"
)

type defaultCouchbaseClusterControl struct {
	recorder record.EventRecorder
}

var _ CouchbaseClusterControl = &defaultCouchbaseClusterControl{}

func NewCouchbaseClusterControl(
	recorder record.EventRecorder,
) CouchbaseClusterControl {
	return &defaultCouchbaseClusterControl{
		recorder: recorder,
	}
}

func (c *defaultCouchbaseClusterControl) SyncCouchbaseCluster(
	cluster v1alpha1.CouchbaseCluster,
) error {
	var err error

	c.recordClusterEvent("sync", cluster, err)
	return nil
}

// recordClusterEvent records an event for verb applied to the CouchbaseCluster. If err is nil the generated event will
// have a reason of apiv1.EventTypeNormal. If err is not nil the generated event will have a reason of apiv1.EventTypeWarning.
func (c *defaultCouchbaseClusterControl) recordClusterEvent(verb string, cluster v1alpha1.CouchbaseCluster, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in CouchbaseCluster %s successful",
			strings.ToLower(verb), cluster.Name)
		c.recorder.Event(&cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Cluster in CouchbaseCluster %s failed error: %s",
			strings.ToLower(verb), cluster.Name, err)
		c.recorder.Event(&cluster, apiv1.EventTypeWarning, reason, message)
	}
}
