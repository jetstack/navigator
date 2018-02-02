package framework

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
)

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	BaseName string

	// A Kubernetes and Service Catalog client
	KubeClientset      kubernetes.Interface
	NavigatorClientset clientset.Interface

	// Namespace in which all test resources should reside
	Namespace *v1.Namespace

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle
}

// NewFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewDefaultFramework(baseName string) *Framework {
	f := &Framework{
		BaseName: baseName,
	}

	BeforeEach(f.BeforeEach)
	// AfterEach(f.AfterEach)

	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	var err error
	By("Creating clientsets")
	f.KubeClientset, err = LoadKubeClientset()
	Expect(err).NotTo(HaveOccurred())
	f.NavigatorClientset, err = LoadNavClientset()
	Expect(err).NotTo(HaveOccurred())

	By("Building a namespace api object")
	f.Namespace, err = CreateKubeNamespace(f.KubeClientset, f.BaseName)
	Expect(err).NotTo(HaveOccurred())
}

// Wrapper function for ginkgo describe.  Adds namespacing.
func NavigatorDescribe(text string, body func()) bool {
	return Describe("[navigator] "+text, body)
}
