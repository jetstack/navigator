package v5

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/glog"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
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
	for _, plugin := range pilot.Spec.Elasticsearch.Plugins {
		if _, ok := installed[plugin.Name]; ok {
			glog.V(4).Infof("Skipping already installed plugin '%s'", plugin.Name)
			continue
		}

		err := p.installPlugin(pilot, plugin.Name)
		if err != nil {
			glog.V(4).Infof("Error installing plugin '%s': %s", plugin.Name, err.Error())
			return err
		}

		glog.V(4).Infof("Successfully installed plugin '%s'", plugin.Name)
	}
	return nil
}

// installPlugin will attempt to install a single plugin on this Elasticsearch
// node.
func (p *Pilot) installPlugin(pilot *v1alpha1.Pilot, plugin string) error {
	cmd := exec.Command(p.Options.ElasticsearchOptions.PluginBinary, "install", plugin)
	cmd.Env = envVars(pilot)
	cmd.Stdout = p.Options.StdOut
	cmd.Stderr = p.Options.StdErr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// getInstalledPlugins will return a list of installed plugins for this Pilot
// by shelling out to the elasticsearch-plugins binary and running 'list'. It
// returns a map containing the plugins name as the key, and an empty struct as
// the value for more efficient indexing.
func (p *Pilot) getInstalledPlugins(pilot *v1alpha1.Pilot) (map[string]struct{}, error) {
	stdout := new(bytes.Buffer)
	cmd := exec.Command(p.Options.ElasticsearchOptions.PluginBinary, "list")
	cmd.Env = envVars(pilot)
	cmd.Stdout = stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	strOutput := stdout.String()
	plugins := strings.Split(strOutput, "\n")
	pluginsMap := make(map[string]struct{})
	for _, plugin := range plugins {
		if len(plugin) == 0 {
			continue
		}
		pluginsMap[plugin] = struct{}{}
	}
	return pluginsMap, nil
}
