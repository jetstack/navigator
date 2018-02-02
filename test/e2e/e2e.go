package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/glog"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/util/logs"

	"github.com/jetstack/navigator/test/e2e/framework"
	"github.com/jetstack/navigator/test/e2e/framework/ginkgowrapper"
)

// There are certain operations we only want to run once per overall test invocation
// (such as deleting old namespaces, or verifying that all system pods are running.
// Because of the way Ginkgo runs tests in parallel, we must use SynchronizedBeforeSuite
// to ensure that these operations only run on the first parallel Ginkgo node.
//
// This function takes two parameters: one function which runs on only the first Ginkgo node,
// returning an opaque byte array, and then a second function which runs on all Ginkgo nodes,
// accepting the byte array.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	// Run only on Ginkgo node 1

	c, err := framework.LoadKubeClientset()
	if err != nil {
		glog.Fatal("Error loading client: ", err)
	}

	// Delete any namespaces except those created by the system. This ensures no
	// lingering resources are left over from a previous test run.
	if framework.TestContext.CleanStart {
		deleted, err := framework.DeleteNamespaces(c, nil, /* deleteFilter */
			[]string{
				metav1.NamespaceSystem,
				metav1.NamespaceDefault,
				metav1.NamespacePublic,
				framework.TestContext.NavigatorNamespace,
			})
		if err != nil {
			framework.Failf("Error deleting orphaned namespaces: %v", err)
		}
		glog.Infof("Waiting for deletion of the following namespaces: %v", deleted)
		if err := framework.WaitForNamespacesDeleted(c, deleted, framework.NamespaceCleanupTimeout); err != nil {
			framework.Failf("Failed to delete orphaned namespaces %v: %v", deleted, err)
		}
	}

	// In large clusters we may get to this point but still have a bunch
	// of nodes without Routes created. Since this would make a node
	// unschedulable, we need to wait until all of them are schedulable.
	framework.ExpectNoError(framework.WaitForAllNodesSchedulable(c, framework.TestContext.NodeSchedulableTimeout))

	// Ensure all pods are running and ready before starting tests (otherwise,
	// cluster infrastructure pods that are being pulled or started can block
	// test pods from running, and tests that ensure all pods are running and
	// ready will fail).
	podStartupTimeout := framework.TestContext.SystemPodsStartupTimeout
	// TODO: In large clusters, we often observe a non-starting pods due to
	// #41007. To avoid those pods preventing the whole test runs (and just
	// wasting the whole run), we allow for some not-ready pods (with the
	// number equal to the number of allowed not-ready nodes).
	if err := framework.WaitForPodsRunningReady(c, metav1.NamespaceSystem, int32(framework.TestContext.MinStartupPods), int32(framework.TestContext.AllowedNotReadyNodes), podStartupTimeout, nil); err != nil {
		framework.DumpAllNamespaceInfo(c, metav1.NamespaceSystem)
		framework.LogFailedContainers(c, metav1.NamespaceSystem, framework.Logf)
		// TODO: Re-enable the kubernetes service tester
		// runKubernetesServiceTestContainer(c, metav1.NamespaceDefault)
		framework.Failf("Error waiting for all pods to be running and ready: %v", err)
	}

	// TODO: re-enable image puller
	// if err := framework.WaitForPodsSuccess(c, metav1.NamespaceSystem, framework.ImagePullerLabels, framework.ImagePrePullingTimeout); err != nil {
	// 	// There is no guarantee that the image pulling will succeed in 3 minutes
	// 	// and we don't even run the image puller on all platforms (including GKE).
	// 	// We wait for it so we get an indication of failures in the logs, and to
	// 	// maximize benefit of image pre-pulling.
	// 	framework.Logf("WARNING: Image pulling pods failed to enter success in %v: %v", framework.ImagePrePullingTimeout, err)
	// }
	// Dump the output of the nethealth containers only once per run
	// if framework.TestContext.DumpLogsOnFailure {
	// 	logFunc := framework.Logf
	// 	if framework.TestContext.ReportDir != "" {
	// 		filePath := path.Join(framework.TestContext.ReportDir, "nethealth.txt")
	// 		file, err := os.Create(filePath)
	// 		if err != nil {
	// 			framework.Logf("Failed to create a file with network health data %v: %v\nPrinting to stdout", filePath, err)
	// 		} else {
	// 			defer file.Close()
	// 			if err = file.Chmod(0644); err != nil {
	// 				framework.Logf("Failed to chmod to 644 of %v: %v", filePath, err)
	// 			}
	// 			logFunc = framework.GetLogToFileFunc(file)
	// 			framework.Logf("Dumping network health container logs from all nodes to file %v", filePath)
	// 		}
	// 	} else {
	// 		framework.Logf("Dumping network health container logs from all nodes...")
	// 	}
	// 	framework.LogContainersInPodsWithLabels(c, metav1.NamespaceSystem, framework.ImagePullerLabels, "nethealth", logFunc)
	// }

	// TODO: wire up version package
	// Log the version of the server and this client.
	// framework.Logf("e2e test version: %s", version.Get().GitVersion)

	dc := c.DiscoveryClient

	serverVersion, serverErr := dc.ServerVersion()
	if serverErr != nil {
		framework.Logf("Unexpected server error retrieving version: %v", serverErr)
	}
	if serverVersion != nil {
		framework.Logf("kube-apiserver version: %s", serverVersion.GitVersion)
	}

	return nil

}, func(data []byte) {
	// Run on all Ginkgo nodes
})

// Similar to SynchornizedBeforeSuite, we want to run some operations only once (such as collecting cluster logs).
// Here, the order of functions is reversed; first, the function which runs everywhere,
// and then the function that only runs on the first Ginkgo node.
var _ = ginkgo.SynchronizedAfterSuite(func() {
	// Run on all Ginkgo nodes
	framework.Logf("Running AfterSuite actions on all node")
	framework.RunCleanupActions()
}, func() {
	// Run only Ginkgo on node 1
	framework.Logf("Running AfterSuite actions on node 1")
	if framework.TestContext.ReportDir != "" {
		// TODO: re-enable core dumps
		// framework.CoreDump(framework.TestContext.ReportDir)
	}
	// TODO: re-enable gathering metrics
	// if framework.TestContext.GatherSuiteMetricsAfterTest {
	// 	if err := gatherTestSuiteMetrics(); err != nil {
	// 		framework.Logf("Error gathering metrics: %v", err)
	// 	}
	// }
})

// RunE2ETests checks configuration parameters (specified through flags) and then runs
// E2E tests using the Ginkgo runner.
// If a "report directory" is specified, one or more JUnit test reports will be
// generated in this directory, and cluster logs will also be saved.
// This function is called on each Ginkgo node in parallel mode.
func RunE2ETests(t *testing.T) {
	runtimeutils.ReallyCrash = true
	logs.InitLogs()
	defer logs.FlushLogs()

	gomega.RegisterFailHandler(ginkgowrapper.Fail)
	// Disable skipped tests unless they are explicitly requested.
	if config.GinkgoConfig.FocusString == "" && config.GinkgoConfig.SkipString == "" {
		config.GinkgoConfig.SkipString = `\[Flaky\]|\[Feature:.+\]`
	}

	// Run tests through the Ginkgo runner with output to console + JUnit for Jenkins
	var r []ginkgo.Reporter
	if framework.TestContext.ReportDir != "" {
		// TODO: we should probably only be trying to create this directory once
		// rather than once-per-Ginkgo-node.
		if err := os.MkdirAll(framework.TestContext.ReportDir, 0755); err != nil {
			glog.Errorf("Failed creating report directory: %v", err)
		} else {
			r = append(r, reporters.NewJUnitReporter(path.Join(framework.TestContext.ReportDir, fmt.Sprintf("junit_%v%02d.xml", framework.TestContext.ReportPrefix, config.GinkgoConfig.ParallelNode))))
		}
	}
	glog.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunId, config.GinkgoConfig.ParallelNode)

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "Navigator e2e suite", r)
}
