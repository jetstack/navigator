package genericpilot

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset/scheme"
	informers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/action"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/hook"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/periodic"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/process"
)

const (
	controllerAgentName = "generic-pilot"
	defaultResyncPeriod = time.Second * 30
)

type Options struct {
	// KubernetesConfig is the config for a Kubernetes API client
	KubernetesConfig *rest.Config
	// NavigatorConfig is the config for a Navigator API client
	NavigatorConfig *rest.Config
	// PilotName is the name of this Pilot
	PilotName string
	// ResyncPeriod for controllers
	ResyncPeriod time.Duration
	// MasterURL for the API server
	MasterURL string
	// KubeConfig
	KubeConfig string

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
	Hooks *hook.Hooks

	SyncFunc func(*v1alpha1.Pilot) error
	// Periodics is a list of Periodic functions to execute on their defined
	// schedule
	Periodics map[string]periodic.Interface
	// Actions is a list of registered Actions for this Pilot. Each Action will
	// usually have a corresponding periodic to update its status.
	Actions map[string]action.Interface
}

func NewDefaultOptions() *Options {
	return &Options{
		ResyncPeriod: defaultResyncPeriod,
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
	var cfg *rest.Config
	var err error
	if o.KubernetesConfig == nil || o.NavigatorConfig == nil {
		cfg, err = clientcmd.BuildConfigFromFlags(o.MasterURL, o.KubeConfig)
		if err != nil {
			return fmt.Errorf("error building api client config: %s", err.Error())
		}
	}
	if o.KubernetesConfig == nil {
		o.KubernetesConfig = cfg
	}
	if o.NavigatorConfig == nil {
		o.NavigatorConfig = cfg
	}
	if o.PilotName == "" {
		o.PilotName, err = os.Hostname()
		if err != nil {
			return fmt.Errorf("error obtaining pilot hostname: %s", err.Error())
		}
	}
	if o.ResyncPeriod == 0 {
		o.ResyncPeriod = defaultResyncPeriod
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
	if o.KubernetesConfig == nil {
		errs = append(errs, fmt.Errorf("kubernetes client config must be specified"))
	}
	if o.NavigatorConfig == nil {
		errs = append(errs, fmt.Errorf("navigator client config must be specified"))
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
	return errs
}

func (o *Options) Pilot() (*GenericPilot, error) {
	kubeClient, err := kubernetes.NewForConfig(o.KubernetesConfig)
	if err != nil {
		return nil, err
	}

	navigatorClient, err := clientset.NewForConfig(o.NavigatorConfig)
	if err != nil {
		return nil, err
	}

	// Create event broadcaster
	// Add navigator types to the default Kubernetes Scheme so Events can be
	// logged for navigator types.
	kubescheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	factory := informers.NewSharedInformerFactory(navigatorClient, o.ResyncPeriod)
	pilotInformer := factory.Navigator().V1alpha1().Pilots()
	pilotInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	go factory.Start(o.StopCh)

	genericPilot := &GenericPilot{
		Options:             *o,
		client:              navigatorClient,
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
	flags.DurationVar(&o.ResyncPeriod, "resync-period", time.Second*30, "Re-sync period for control loops operated by the pilot")
	flags.StringVar(&o.KubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flags.StringVar(&o.MasterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
