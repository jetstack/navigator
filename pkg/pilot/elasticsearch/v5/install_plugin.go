package v5

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

// InstallPlugins will install the plugins listed on the Pilot resource. It
// will not uninstall any plugins that are already installed, but are not
// listed on the Pilot resource.
func (p *Pilot) InstallPlugins(pilot *v1alpha1.Pilot) error {
	installed, err := p.getInstalledPlugins(pilot)
	if err != nil {
		return fmt.Errorf("error listing installed plugins: %s", err.Error())
	}
	glog.V(4).Infof("There are %d plugins already installed: %v", len(installed), installed)
	for _, plugin := range p.Options.ElasticsearchOptions.Plugins {
		if installed.Has(plugin) {
			glog.V(4).Infof("Skipping already installed plugin '%s'", plugin)
			continue
		}

		err := p.installPlugin(pilot, plugin)
		if err != nil {
			glog.V(4).Infof("Error installing plugin '%s': %s", plugin, err.Error())
			return err
		}

		glog.V(4).Infof("Successfully installed plugin '%s'", plugin)
	}
	return nil
}

// installPlugin will attempt to install a single plugin on this Elasticsearch
// node.
func (p *Pilot) installPlugin(pilot *v1alpha1.Pilot, plugin string) error {
	cmd := exec.Command(p.Options.ElasticsearchOptions.PluginBinary, "install", plugin)
	cmd.Env = p.env().Strings()
	cmd.Stdout = p.Options.StdOut
	cmd.Stderr = p.Options.StdErr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// getInstalledPlugins will return a list of installed plugins for this Pilot
// by shelling out to the elasticsearch-plugins binary and running 'list'.
func (p *Pilot) getInstalledPlugins(pilot *v1alpha1.Pilot) (sets.String, error) {
	stdout := new(bytes.Buffer)
	cmd := exec.Command(p.Options.ElasticsearchOptions.PluginBinary, "list")
	cmd.Env = p.env().Strings()
	cmd.Stdout = stdout
	cmd.Stderr = p.Options.StdErr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	strOutput := stdout.String()
	pluginsSlice := strings.Split(strOutput, "\n")
	return sets.NewString(pluginsSlice...), nil
}
