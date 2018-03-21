package elasticsearch

import (
	. "github.com/onsi/ginkgo"

	"k8s.io/client-go/kubernetes"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	"github.com/jetstack/navigator/test/e2e/framework"
)

var _ = Describe("Kubectl tests", func() {
	f := framework.NewDefaultFramework("elasticsearch-shortname")
	var ns string
	var kubeClient kubernetes.Interface
	var navClient clientset.Interface

	BeforeEach(func() {
		kubeClient = f.KubeClientset
		navClient = f.NavigatorClientset
		ns = f.Namespace.Name
	})

	framework.NavigatorDescribe("Kubectl functionality [ElasticsearchKubectl]", func() {
		AfterEach(func() {
			if CurrentGinkgoTestDescription().Failed {
				framework.DumpDebugInfo(kubeClient, ns)
			}
		})

		It("should support the 'esc' shortname with kubectl [ElasticsearchShortName]", func() {
			err := framework.KubectlCmd("get", "esc").Run()
			framework.ExpectNoError(err, "got an error whilst listing esc (ElasticsearchClusters)")
		})
	})
})
