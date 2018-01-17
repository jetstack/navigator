package v5

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"gopkg.in/olivere/elastic.v5"
	"k8s.io/client-go/tools/cache"

	clientset "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	listersv1alpha1 "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot"
	"github.com/jetstack/navigator/pkg/pilot/genericpilot/hook"
)

const (
	localESClientURL = "http://127.0.0.1:9200"
)

type Pilot struct {
	Options *PilotOptions

	// navigator clientset
	navigatorClient clientset.Interface

	pilotLister             listersv1alpha1.PilotLister
	pilotInformerSynced     cache.InformerSynced
	esClusterLister         listersv1alpha1.ElasticsearchClusterLister
	esClusterInformerSynced cache.InformerSynced

	// client for accessing the elasticsearch API locally
	localESClient *elastic.Client

	// a reference to the GenericPilot for this Pilot
	genericPilot *genericpilot.GenericPilot

	lastNodeStatsUpdate time.Time
}

func NewPilot(opts *PilotOptions) (*Pilot, error) {
	pilotInformer := opts.sharedInformerFactory.Navigator().V1alpha1().Pilots()
	esClusterInformer := opts.sharedInformerFactory.Navigator().V1alpha1().ElasticsearchClusters()

	p := &Pilot{
		Options:                 opts,
		navigatorClient:         opts.navigatorClientset,
		pilotLister:             pilotInformer.Lister(),
		pilotInformerSynced:     pilotInformer.Informer().HasSynced,
		esClusterLister:         esClusterInformer.Lister(),
		esClusterInformerSynced: esClusterInformer.Informer().HasSynced,
	}

	// Setup a gofunc to keep attempting to create an API client
	go func() {
		for {
			cl, err := elastic.NewClient(elastic.SetHttpClient(http.DefaultClient), elastic.SetURL(localESClientURL))
			if err == nil {
				p.localESClient = cl
				break
			}
			glog.Errorf("Error creating elasticsearch api client: %s", err.Error())
		}
	}()

	return p, nil
}

func (p *Pilot) WaitForCacheSync(stopCh <-chan struct{}) error {
	if !cache.WaitForCacheSync(stopCh, p.pilotInformerSynced, p.esClusterInformerSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	return nil
}

func (p *Pilot) Hooks() *hook.Hooks {
	return &hook.Hooks{
		PreStart: []hook.Interface{
			hook.New("WriteConfig", p.WriteConfig),
			hook.New("InstallPlugins", p.InstallPlugins),
		},
	}
}
