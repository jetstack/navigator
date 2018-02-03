package framework

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	esutil "github.com/jetstack/navigator/pkg/controllers/elasticsearch/util"
)

const (
	// Poll interval for ElasticsearchCluster tests
	ElasticsearchClusterPoll = 10 * time.Second
	// Timeout interval for ElasticsearchCluster operations
	ElasticsearchClusterTimeout = 10 * time.Minute
)

// DeleteAllElasticsearchClusters deletes all ElasticsearchCluster API Objects in Namespace ns.
func DeleteAllElasticsearchClusters(c clientset.Interface, ns string) {
	esList, err := c.NavigatorV1alpha1().ElasticsearchClusters(ns).List(metav1.ListOptions{LabelSelector: labels.Everything().String()})
	ExpectNoError(err)

	errList := []string{}
	for i := range esList.Items {
		es := &esList.Items[i]
		Logf("Deleting elasticsearchcluster %v", es.Name)
		// Use OrphanDependents=false so it's deleted synchronously.
		// We already made sure the Pods are gone inside Scale().
		if err := c.NavigatorV1alpha1().ElasticsearchClusters(es.Namespace).Delete(es.Name, &metav1.DeleteOptions{OrphanDependents: new(bool)}); err != nil {
			errList = append(errList, fmt.Sprintf("%v", err))
		}
	}

	if len(errList) != 0 {
		ExpectNoError(fmt.Errorf("%v", strings.Join(errList, "\n")))
	}
}

// ElasticsearchTester is a struct that contains utility methods for testing
// ElasticsearchCluster related functionality. It uses a clientset.Interface to
// communicate with the API server.
type ElasticsearchTester struct {
	kubeClient kubernetes.Interface
	navClient  clientset.Interface
}

func NewElasticsearchTester(kubeClient kubernetes.Interface, navClient clientset.Interface) *ElasticsearchTester {
	return &ElasticsearchTester{kubeClient, navClient}
}

func (e *ElasticsearchTester) CreateClusterAndWaitForReady(es *v1alpha1.ElasticsearchCluster) *v1alpha1.ElasticsearchCluster {
	var err error
	By("Creating elasticsearchCluster " + es.Name + " in namespace " + es.Namespace)
	es, err = e.navClient.NavigatorV1alpha1().ElasticsearchClusters(es.Namespace).Create(es)
	Expect(err).NotTo(HaveOccurred())
	By("Waiting for CreateNodePool event")
	WaitForElasticsearchClusterEvent(e.kubeClient, es.Name, es.Namespace, "CreateNodePool")
	By("Waiting for cluster pods to be ready")
	e.WaitForAllReady(es)
	return es
}

func (e *ElasticsearchTester) ScaleNodePool(es *v1alpha1.ElasticsearchCluster, pool string, replicas int32) {
	var err error
	By("Retrieving an up to date copy of the cluster resource")
	es, err = e.navClient.NavigatorV1alpha1().ElasticsearchClusters(es.Namespace).Get(es.Name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	found := false
	for i, np := range es.Spec.NodePools {
		if np.Name == pool {
			np.Replicas = replicas
			es.Spec.NodePools[i] = np
			found = true
			break
		}
	}
	if !found {
		Failf("Node pool %q not found in ElasticsearchCluster %q", pool, es.Name)
	}
	By("Scaling node pool " + es.Name + "/" + pool + " in namespace " + es.Namespace)
	es, err = e.navClient.NavigatorV1alpha1().ElasticsearchClusters(es.Namespace).Update(es)
	Expect(err).NotTo(HaveOccurred())
	By("Waiting for Scale event")
	WaitForElasticsearchClusterEvent(e.kubeClient, es.Name, es.Namespace, "Scale")
	By("Waiting for cluster pods to be ready")
	e.WaitForAllReady(es)
}

func (e *ElasticsearchTester) WaitForHealth(es *v1alpha1.ElasticsearchCluster, expectedHealth v1alpha1.ElasticsearchClusterHealth) {
	pollErr := wait.PollImmediate(ElasticsearchClusterPoll, ElasticsearchClusterTimeout,
		func() (bool, error) {
			latestEs, err := e.navClient.NavigatorV1alpha1().ElasticsearchClusters(es.Namespace).Get(es.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			Logf("Waiting for ElasticsearchCluster %q to be in state %q. Current: %q", es.Name, expectedHealth, latestEs.Status.Health)
			if latestEs.Status.Health == expectedHealth {
				return true, nil
			}
			return false, nil
		})
	if pollErr != nil {
		Failf("Failed waiting for elasticsearchcluster to be in %s state: %v", expectedHealth, pollErr)
	}
}

func (e *ElasticsearchTester) WaitForAllReady(es *v1alpha1.ElasticsearchCluster) {
	var i int32
	for _, np := range es.Spec.NodePools {
		for i = 0; i < np.Replicas; i++ {
			Logf("Waiting for stateful pod at index %v to enter Running", i)
			e.WaitForNodePoolReady(es, &np)
		}
	}
}

func (e *ElasticsearchTester) WaitForNodePoolReady(es *v1alpha1.ElasticsearchCluster, np *v1alpha1.ElasticsearchClusterNodePool) {
	ssName := esutil.NodePoolResourceName(es, np)
	tester := NewStatefulSetTester(e.kubeClient)
	ss, err := e.kubeClient.AppsV1beta1().StatefulSets(es.Namespace).Get(ssName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	tester.WaitForRunning(*ss.Spec.Replicas, *ss.Spec.Replicas, ss)
}

func DefaultElasticsearchPilotImageSpec() v1alpha1.ImageSpec {
	return v1alpha1.ImageSpec{
		Repository: TestContext.ESPilotImageRepo,
		Tag:        TestContext.ESPilotImageTag,
		PullPolicy: core.PullNever,
	}
}

func DefaultElasticsearchSysctls() []string {
	return []string{"vm.max_map_count=262144"}
}

func DefaultElasticsearchNodeResources() core.ResourceRequirements {
	return core.ResourceRequirements{
		Requests: core.ResourceList{
			core.ResourceCPU:    resource.MustParse("1"),
			core.ResourceMemory: resource.MustParse("2Gi"),
		},
		Limits: core.ResourceList{
			core.ResourceCPU:    resource.MustParse("1500m"),
			core.ResourceMemory: resource.MustParse("3500Mi"),
		},
	}
}

func GetElasticsearchClusterEvents(c kubernetes.Interface, name, ns string) []core.Event {
	selector := fields.Set{
		"involvedObject.kind":      "ElasticsearchCluster",
		"involvedObject.name":      name,
		"involvedObject.namespace": ns,
		"source":                   "navigator-controller",
	}.AsSelector().String()
	options := metav1.ListOptions{FieldSelector: selector}
	events, err := c.CoreV1().Events(ns).List(options)
	if err != nil {
		Logf("Unexpected error retrieving node events %v", err)
		return []core.Event{}
	}
	return events.Items
}

func WaitForElasticsearchClusterEvent(c kubernetes.Interface, name, ns, reason string) {
	pollErr := wait.PollImmediate(ElasticsearchClusterPoll, ElasticsearchClusterTimeout,
		func() (bool, error) {
			evts := GetElasticsearchClusterEvents(c, name, ns)
			for _, e := range evts {
				if e.Reason == reason {
					return true, nil
				}
			}
			return false, nil
		})
	if pollErr != nil {
		Failf("Failed waiting for %q event for elasticsearchcluster %q", reason, name)
	}
}
