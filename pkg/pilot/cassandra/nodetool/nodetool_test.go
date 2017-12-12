package nodetool_test

import (
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/integration/framework"
)

type Framework struct {
}

func NewFramework() *Framework {
	return &Framework{}
}

func (f *Framework) StartCassandraCluster() {

}

func TestNodeTool(t *testing.T) {
	masterConfig := framework.NewIntegrationTestMasterConfig()
	_, server, closeFn := framework.RunAMaster(masterConfig)

	defer closeFn()
	config := restclient.Config{Host: server.URL}
	client, err := kubernetes.NewForConfig(&config)
	if err != nil {
		t.Fatalf("Error in creating clientset: %v", err)
	}
	t.Log(client)
	// f := NewFramework(Framework)
	// f.StartCassandraCluster()
	// nt := nodetool.New()
	// nt.Status()
	// t.Log(nt)
	// t.Fatal("foo")
}
