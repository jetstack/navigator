package elasticsearch

import (
	. "github.com/onsi/ginkgo"

	"k8s.io/client-go/kubernetes"

	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	"github.com/jetstack/navigator/test/e2e/framework"
)

var _ = Describe("Scale tests", func() {
	f := framework.NewDefaultFramework("elasticsearch-scale")
	var ns string
	var kubeClient kubernetes.Interface
	var navClient clientset.Interface

	BeforeEach(func() {
		kubeClient = f.KubeClientset
		navClient = f.NavigatorClientset
		ns = f.Namespace.Name
	})

	framework.NavigatorDescribe("Elasticsearch scaling functionality [ElasticsearchScale]", func() {
		clusterName := "test"

		AfterEach(func() {
			if CurrentGinkgoTestDescription().Failed {
				framework.DumpDebugInfo(kubeClient, ns)
			}
			framework.Logf("Deleting all elasticsearchClusters in ns %v", ns)
			framework.DeleteAllElasticsearchClusters(navClient, ns)
			framework.DeleteAllStatefulSets(kubeClient, ns)
			framework.WaitForNoPodsInNamespace(kubeClient, ns, framework.NamespaceCleanupTimeout)
		})

		It("should scale up a single node cluster to make it turn green from yellow", func() {
			nodePoolName := "mixed"
			cluster := generate.Cluster(generate.ClusterConfig{
				Name:      clusterName,
				Namespace: ns,
				Version:   "5.6.2",
				ClusterConfig: v1alpha1.NavigatorClusterConfig{
					PilotImage: framework.DefaultElasticsearchPilotImageSpec(),
					Sysctls:    framework.DefaultElasticsearchSysctls(),
				},
				NodePools: []v1alpha1.ElasticsearchClusterNodePool{
					{
						Name:      nodePoolName,
						Replicas:  1,
						Resources: framework.DefaultElasticsearchNodeResources(),
						Roles: []v1alpha1.ElasticsearchClusterRole{
							v1alpha1.ElasticsearchRoleData,
							v1alpha1.ElasticsearchRoleIngest,
							v1alpha1.ElasticsearchRoleMaster,
						},
					},
				},
			})
			tester := framework.NewElasticsearchTester(kubeClient, navClient)
			cluster = tester.CreateClusterAndWaitForReady(cluster)
			By("Waiting for the cluster to be in a Yellow state")
			tester.WaitForHealth(cluster, v1alpha1.ElasticsearchClusterHealthYellow)
			By("Scaling up the '" + nodePoolName + "' node pool")
			tester.ScaleNodePool(cluster, nodePoolName, 2)
			By("Waiting for the cluster to be in a Green state")
			tester.WaitForHealth(cluster, v1alpha1.ElasticsearchClusterHealthGreen)
		})
	})
})
