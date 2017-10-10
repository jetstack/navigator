package v5

import (
	"github.com/golang/glog"
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

func (p *Pilot) InstallPlugins(pilot *v1alpha1.Pilot) error {
	glog.V(4).Infof("Installing plugins")
	return nil
}
