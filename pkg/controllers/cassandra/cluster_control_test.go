package cassandra_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func CassandraClusterSummary(c *v1alpha1.CassandraCluster) string {
	return fmt.Sprintf(
		"%s/%s {Spec: %s, Status: %s}",
		c.Namespace, c.Name,
		CassandraClusterSpecSummary(c),
		CassandraClusterStatusSummary(c),
	)
}

func CassandraClusterSpecSummary(c *v1alpha1.CassandraCluster) string {
	nodepools := make([]string, len(c.Spec.NodePools))
	for i, np := range c.Spec.NodePools {
		nodepools[i] = fmt.Sprintf("%s:%d", np.Name, np.Replicas)
	}
	return fmt.Sprintf(
		"{nodepools: %s}",
		strings.Join(nodepools, ", "),
	)
}

func CassandraClusterStatusSummary(c *v1alpha1.CassandraCluster) string {
	nodepools := make([]string, len(c.Status.NodePools))
	i := 0
	for title, nps := range c.Status.NodePools {
		nodepools[i] = fmt.Sprintf("%s:%d", title, nps.ReadyReplicas)
		i++
	}
	return fmt.Sprintf(
		"{nodepools: %s}", strings.Join(nodepools, ", "),
	)
}

func TestNextAction(t *testing.T) {
	f := func(c *v1alpha1.CassandraCluster) bool {
		t.Log(CassandraClusterSummary(c))
		a := cassandra.NextAction(c)
		if a != nil {
			t.Log("Action:", a.Name())
		} else {
			t.Log("No action")
		}
		switch action := a.(type) {
		case *actions.CreateNodePool:
			_, found := c.Status.NodePools[action.NodePool.Name]
			if found {
				t.Errorf("Unexpected attempt to create a nodepool when there's an existing status")
				return false
			}
		case *actions.ScaleOut:
			nps, found := c.Status.NodePools[action.NodePool.Name]
			if !found {
				t.Errorf("Unexpected attempt to scale up a nodepool without a status")
				return false
			}
			if action.NodePool.Replicas <= nps.ReadyReplicas {
				t.Errorf("Unexpected attempt to scale up a nodepool with >= ready replicas")
				return false
			}
		}
		return true
	}
	config := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, rnd *rand.Rand) {
			cluster := &v1alpha1.CassandraCluster{}
			cluster.SetName("cluster1")
			cluster.SetNamespace("ns1")
			casstesting.FuzzCassandraClusterNodePools(cluster, rnd, 0)
			values[0] = reflect.ValueOf(cluster)
		},
	}
	err := quick.Check(f, config)
	if err != nil {
		t.Errorf("quick check failure: %#v", err)
	}
}
