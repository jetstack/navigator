package framework

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	apps "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// Poll interval for StatefulSet tests
	StatefulSetPoll = 10 * time.Second
	// Timeout interval for StatefulSet operations
	StatefulSetTimeout = 10 * time.Minute
	// Timeout for stateful pods to change state
	StatefulPodTimeout = 5 * time.Minute
)

type StatefulSetTester struct {
	c kubernetes.Interface
}

// NewStatefulSetTester creates a StatefulSetTester that uses c to interact with the API server.
func NewStatefulSetTester(c kubernetes.Interface) *StatefulSetTester {
	return &StatefulSetTester{c}
}

// SortStatefulPods sorts pods by their ordinals
func (s *StatefulSetTester) SortStatefulPods(pods *v1.PodList) {
	sort.Sort(statefulPodsByOrdinal(pods.Items))
}

// GetPodList gets the current Pods in ss.
func (s *StatefulSetTester) GetPodList(ss *apps.StatefulSet) *v1.PodList {
	selector, err := metav1.LabelSelectorAsSelector(ss.Spec.Selector)
	ExpectNoError(err)
	podList, err := s.c.CoreV1().Pods(ss.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	ExpectNoError(err)
	return podList
}

// WaitForRunning waits for numPodsRunning in ss to be Running and for the first
// numPodsReady ordinals to be Ready.
func (s *StatefulSetTester) WaitForRunning(numPodsRunning, numPodsReady int32, ss *apps.StatefulSet) {
	pollErr := wait.PollImmediate(StatefulSetPoll, StatefulSetTimeout,
		func() (bool, error) {
			podList := s.GetPodList(ss)
			s.SortStatefulPods(podList)
			if int32(len(podList.Items)) < numPodsRunning {
				Logf("Found %d stateful pods, waiting for %d", len(podList.Items), numPodsRunning)
				return false, nil
			}
			if int32(len(podList.Items)) > numPodsRunning {
				return false, fmt.Errorf("Too many pods scheduled, expected %d got %d", numPodsRunning, len(podList.Items))
			}
			for _, p := range podList.Items {
				shouldBeReady := getStatefulPodOrdinal(&p) < int(numPodsReady)
				isReady := IsPodReady(&p)
				desiredReadiness := shouldBeReady == isReady
				Logf("Waiting for pod %v to enter %v - Ready=%v, currently %v - Ready=%v", p.Name, v1.PodRunning, shouldBeReady, p.Status.Phase, isReady)
				if p.Status.Phase != v1.PodRunning || !desiredReadiness {
					return false, nil
				}
			}
			return true, nil
		})
	if pollErr != nil {
		Failf("Failed waiting for pods to enter running: %v", pollErr)
	}
}

// DeleteAllStatefulSets deletes all StatefulSet API Objects in Namespace ns.
func DeleteAllStatefulSets(c kubernetes.Interface, ns string) {
	esList, err := c.AppsV1beta1().StatefulSets(ns).List(metav1.ListOptions{LabelSelector: labels.Everything().String()})
	ExpectNoError(err)

	errList := []string{}
	for i := range esList.Items {
		es := &esList.Items[i]
		Logf("Deleting statefulset %v", es.Name)
		// Use OrphanDependents=false so it's deleted synchronously.
		// We already made sure the Pods are gone inside Scale().
		if err := c.AppsV1beta1().StatefulSets(ns).Delete(es.Name, &metav1.DeleteOptions{OrphanDependents: new(bool)}); err != nil {
			errList = append(errList, fmt.Sprintf("%v", err))
		}
	}

	if len(errList) != 0 {
		ExpectNoError(fmt.Errorf("%v", strings.Join(errList, "\n")))
	}
}

var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

func getStatefulPodOrdinal(pod *v1.Pod) int {
	ordinal := -1
	subMatches := statefulPodRegex.FindStringSubmatch(pod.Name)
	if len(subMatches) < 3 {
		return ordinal
	}
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return ordinal
}

type statefulPodsByOrdinal []v1.Pod

func (sp statefulPodsByOrdinal) Len() int {
	return len(sp)
}

func (sp statefulPodsByOrdinal) Swap(i, j int) {
	sp[i], sp[j] = sp[j], sp[i]
}

func (sp statefulPodsByOrdinal) Less(i, j int) bool {
	return getStatefulPodOrdinal(&sp[i]) < getStatefulPodOrdinal(&sp[j])
}
