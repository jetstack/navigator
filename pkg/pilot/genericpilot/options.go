package genericpilot

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	"github.com/jetstack/navigator/pkg/client/clientset/versioned/scheme"
	informers "github.com/jetstack/navigator/pkg/client/informers/externalversions"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/controller"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/leaderelection"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/probe"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/processmanager"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/signals"
)

const (
	controllerAgentName = "generic-pilot"
)

type Options struct {
	// KubernetesClient is the kubernetes clientset used to talk to the
	// apiserver
	KubernetesClient kubernetes.Interface
	// NavigatorClient is the clientset used to talk to the navigator apiserver
	NavigatorClient clientset.Interface
	// SharedInformerFactory provides a shared cache of informers
	SharedInformerFactory informers.SharedInformerFactory

	// KubeSharedInformerFactory provides a shared cache of informers
	KubeSharedInformerFactory kubeinformers.SharedInformerFactory

	// PilotName is the name of this Pilot
	PilotName string
	// PilotNamespace is the namespace the corresponding Pilot exists within
	PilotNamespace          string
	LeaderElectionConfigMap string

	// CmdFunc returns an *exec.Cmd for a given Pilot resource for the pilot
	CmdFunc func(*v1alpha1.Pilot) (*exec.Cmd, error)
	// Signals contains a genericpilot->os.Signal translation
	Signals processmanager.Signals
	// Stdout to be used when creating the application process
	Stdout *os.File
	// Stderr to be used when creating the application process
	Stderr *os.File
	// StopCh signals that the Pilot should shut down when closed
	StopCh <-chan struct{}

	// Hooks to be run during the lifecycle of the application
	Hooks          *hook.Hooks
	ReadinessProbe probe.Check
	LivenessProbe  probe.Check

	SyncFunc              func(*v1alpha1.Pilot) error
	LeaderElectedSyncFunc func(*v1alpha1.Pilot) error
}

func NewDefaultOptions() *Options {
	return &Options{
		Signals: processmanager.Signals{
			Stop: syscall.SIGTERM,
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (o *Options) Complete() error {
	var err error
	if o.PilotName == "" {
		o.PilotName, err = os.Hostname()
		if err != nil {
			return fmt.Errorf("error obtaining pilot hostname: %s", err.Error())
		}
	}
	if o.StopCh == nil {
		o.StopCh = signals.SetupSignalHandler()
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
	if o.Signals.Stop == nil {
		o.Signals.Stop = syscall.SIGTERM
	}
	if o.Hooks == nil {
		o.Hooks = &hook.Hooks{}
	}
	return nil
}

func (o *Options) Validate() []error {
	var errs []error
	if o.KubernetesClient == nil {
		errs = append(errs, fmt.Errorf("kubernetes client must be specified"))
	}
	if o.NavigatorClient == nil {
		errs = append(errs, fmt.Errorf("navigator client must be specified"))
	}
	if o.SharedInformerFactory == nil {
		errs = append(errs, fmt.Errorf("shared informer factory must be specified"))
	}
	if o.KubeSharedInformerFactory == nil {
		errs = append(errs, fmt.Errorf("kube shared informer factory must be specified"))
	}
	if o.CmdFunc == nil {
		errs = append(errs, fmt.Errorf("cmd func must be specified"))
	}
	if o.Stderr == nil {
		errs = append(errs, fmt.Errorf("stderr must be specified"))
	}
	if o.Stdout == nil {
		errs = append(errs, fmt.Errorf("stdout must be specified"))
	}
	if o.PilotName == "" {
		errs = append(errs, fmt.Errorf("pilot name must be specified"))
	}
	if o.SyncFunc == nil {
		errs = append(errs, fmt.Errorf("syncfunc must be specified"))
	}
	if o.Hooks == nil {
		errs = append(errs, fmt.Errorf("hooks must not be nil"))
	}
	if o.PilotNamespace == "" {
		errs = append(errs, fmt.Errorf("pilot namespace must be specified"))
	}
	return errs
}

func (o *Options) Pilot() (*GenericPilot, error) {
	// Create event broadcaster
	// Add navigator types to the default Kubernetes Scheme so Events can be
	// logged for navigator types.
	kubescheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: o.KubernetesClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	// TODO: use a filtered informer that only watches pilot-namespace
	pilotInformer := o.SharedInformerFactory.Navigator().V1alpha1().Pilots()
	genericPilot := &GenericPilot{
		Options:     *o,
		client:      o.NavigatorClient,
		pilotLister: pilotInformer.Lister(),
		recorder:    recorder,
		elector: &leaderelection.Elector{
			LockMeta: metav1.ObjectMeta{
				Name:      o.LeaderElectionConfigMap,
				Namespace: o.PilotNamespace,
			},
			Client:   o.KubernetesClient,
			Recorder: recorder,
		},
	}
	genericPilot.controller = controller.NewController(controller.Options{
		PilotName:      o.PilotName,
		PilotNamespace: o.PilotNamespace,
		SyncFunc:       genericPilot.syncPilot,
		KubeClientset:  o.KubernetesClient,
		Clientset:      o.NavigatorClient,
		PilotInformer:  pilotInformer,
		State: &controllers.State{
			PilotLister:       o.SharedInformerFactory.Navigator().V1alpha1().Pilots().Lister(),
			PodLister:         o.KubeSharedInformerFactory.Core().V1().Pods().Lister(),
			StatefulSetLister: o.KubeSharedInformerFactory.Apps().V1beta1().StatefulSets().Lister(),
		},
	})

	return genericPilot, nil
}

func (o *Options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.PilotName, "pilot-name", "", "The name of this Pilot. If not specified, an auto-detected name will be used.")
	flags.StringVar(&o.PilotNamespace, "pilot-namespace", "", "The namespace the corresponding Pilot resource for this Pilot exists within.")
	flags.StringVar(&o.LeaderElectionConfigMap, "leader-election-config-map", "", "The  name of the ConfigMap to use for leader election")
}
