package v5

import (
	"fmt"

	"k8s.io/client-go/tools/cache"

	"github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	listersv1alpha1 "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/v1alpha1"
	"github.com/jetstack-experimental/navigator/pkg/pilot/genericpilot/hook"
)

type Pilot struct {
	Options *PilotOptions

	navigatorClient     clientset.Interface
	pilotLister         listersv1alpha1.PilotLister
	pilotInformerSynced cache.InformerSynced

	esClusterLister         listersv1alpha1.ElasticsearchClusterLister
	esClusterInformerSynced cache.InformerSynced
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
