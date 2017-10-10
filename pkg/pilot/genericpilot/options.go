package genericpilot

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/scheme"
	informers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/probe"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/process"
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

	// PilotName is the name of this Pilot
	PilotName string
	// PilotNamespace is the namespace the corresponding Pilot exists within
	PilotNamespace string

	// CmdFunc returns an *exec.Cmd for a given Pilot resource for the pilot
	CmdFunc func(*v1alpha1.Pilot) (*exec.Cmd, error)
	// Signals contains a genericpilot->os.Signal translation
	Signals process.Signals
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

	SyncFunc func(*v1alpha1.Pilot) error
}

func NewDefaultOptions() *Options {
	return &Options{
		Signals: process.Signals{
			Stop:      syscall.SIGTERM,
			Terminate: syscall.SIGKILL,
			Reload:    syscall.SIGHUP,
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
		o.StopCh = SetupSignalHandler()
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
	if o.Signals.Terminate == nil {
		o.Signals.Stop = syscall.SIGKILL
	}
	if o.Signals.Reload == nil {
		o.Signals.Stop = syscall.SIGHUP
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
	pilotInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	genericPilot := &GenericPilot{
		Options:             *o,
		client:              o.NavigatorClient,
		pilotLister:         pilotInformer.Lister(),
		pilotInformerSynced: pilotInformer.Informer().HasSynced,
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pilots"),
		recorder:            recorder,
	}

	pilotInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: genericPilot.enqueuePilot,
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				genericPilot.enqueuePilot(new)
			}
		},
		DeleteFunc: genericPilot.enqueuePilot,
	})

	return genericPilot, nil
}

func (o *Options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.PilotName, "pilot-name", "", "The name of this Pilot. If not specified, an auto-detected name will be used.")
	flags.StringVar(&o.PilotNamespace, "pilot-namespace", "", "The namespace the corresponding Pilot resource for this Pilot exists within.")
}
