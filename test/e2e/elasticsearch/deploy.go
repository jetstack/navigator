package elasticsearch

import (
	. "github.com/onsi/ginkgo"

	"k8s.io/client-go/kubernetes"

	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	"github.com/jetstack/navigator/test/e2e/framework"
)

var _ = Describe("Deployment tests", func() {
	f := framework.NewDefaultFramework("elasticsearch-deployment")
	var ns string
	var kubeClient kubernetes.Interface
	var navClient clientset.Interface

	BeforeEach(func() {
		kubeClient = f.KubeClientset
		navClient = f.NavigatorClientset
		ns = f.Namespace.Name
	})

	framework.NavigatorDescribe("Basic Elasticsearch deployment functionality [ElasticsearchDeployBasic]", func() {
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

		It("should deploy a single node elasticsearch cluster", func() {
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
						Name:      "mixed",
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
		})

		It("should deploy a 3 node, 2 node pool elasticsearch cluster", func() {
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
						Name:      "master",
						Replicas:  1,
						Resources: framework.DefaultElasticsearchNodeResources(),
						Roles: []v1alpha1.ElasticsearchClusterRole{
							v1alpha1.ElasticsearchRoleMaster,
						},
					},
					{
						Name:      "mixed",
						Replicas:  2,
						Resources: framework.DefaultElasticsearchNodeResources(),
						Roles: []v1alpha1.ElasticsearchClusterRole{
							v1alpha1.ElasticsearchRoleData,
							v1alpha1.ElasticsearchRoleIngest,
						},
					},
				},
			})
			tester := framework.NewElasticsearchTester(kubeClient, navClient)
			cluster = tester.CreateClusterAndWaitForReady(cluster)
			By("Waiting for the cluster to be in a Green state")
			tester.WaitForHealth(cluster, v1alpha1.ElasticsearchClusterHealthGreen)
			// TODO: ensure documents are being written
		})
	})
})
