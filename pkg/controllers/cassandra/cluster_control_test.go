package cassandra_test

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/kr/pretty"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestNextAction(t *testing.T) {
	f := func(c *v1alpha1.CassandraCluster) (ret bool) {
		defer func() {
			if !ret {
				t.Logf(pretty.Sprint(c))
			}
		}()
		a := cassandra.NextAction(c)
		switch action := a.(type) {
		case *actions.CreateNodePool:
			_, found := c.Status.NodePools[action.NodePool.Name]
			if found {
				t.Errorf("Unexpected attempt to create a nodepool when there's an existing status")
				return false
			}
		case *actions.UpdateVersion:
			nps, found := c.Status.NodePools[action.NodePool.Name]
			if !found {
				t.Errorf("Unexpected updateversion before status reported")
				return false
			}
			if nps.Version == nil {
				t.Errorf("Unexpected updateversion before version reported")
				return false
			}
			if nps.Version.Major != c.Spec.Version.Major {
				t.Errorf("Unexpected updateversion for major version change")
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
		MaxCount: 1000,
		Values: func(values []reflect.Value, rnd *rand.Rand) {
			cluster := &v1alpha1.CassandraCluster{}
			cluster.SetName("cluster1")
			cluster.SetNamespace("ns1")
			casstesting.FuzzCassandraCluster(cluster, rnd, 0)
			values[0] = reflect.ValueOf(cluster)
		},
	}
	err := quick.Check(f, config)
	if err != nil {
		t.Errorf("quick check failure: %#v", err)
	}
}
