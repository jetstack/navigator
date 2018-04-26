/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/admission"
	admissionmetrics "k8s.io/apiserver/pkg/admission/metrics"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/jetstack/navigator/pkg/api"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/apiserver"
	navigatorinitializer "github.com/jetstack/navigator/pkg/apiserver/admission"
	clientset "github.com/jetstack/navigator/pkg/client/clientset/internalversion"
	informers "github.com/jetstack/navigator/pkg/client/informers/internalversion"
)

const defaultEtcdPathPrefix = "/registry/navigator.jetstack.io"

type NavigatorServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	Admission          *genericoptions.AdmissionOptions

	StandaloneMode bool
	StdOut         io.Writer
	StdErr         io.Writer
}

func NewNavigatorServerOptions(out, errOut io.Writer) *NavigatorServerOptions {
	o := &NavigatorServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, apiserver.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion)),
		Admission:          genericoptions.NewAdmissionOptions(),

		StdOut: out,
		StdErr: errOut,
	}

	return o
}

// NewCommandStartNavigatorServer provides a CLI handler for the 'navigator-apiserver' command
func NewCommandStartNavigatorServer(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := NewNavigatorServerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:   "navigator-apiserver",
		Short: "Launch a Navigator API server",
		Long: `
Launch a Navigator API server.

Navigator is a Kubernetes extension for managing common stateful services on Kubernetes.
Documentation is available at https://navigator-dbaas.readthedocs.io.
`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunNavigatorServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.AddFlags(flags)
	o.RecommendedOptions.AddFlags(flags)
	o.Admission.AddFlags(flags)

	return cmd
}

func (o NavigatorServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	errors = append(errors, o.Admission.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *NavigatorServerOptions) Complete() error {
	return nil
}

// AddFlags adds flags related to Navigator to the specified FlagSet
func (o *NavigatorServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.StandaloneMode, "standalone-mode", false, ""+
		"Standalone mode runs the APIServer in a mode that doesn't require a "+
		"connection to a core Kubernetes API server. For example, admission "+
		"control is disabled in standalone mode.")
}

func (o NavigatorServerOptions) Config() (*apiserver.Config, error) {
	// register admission plugins
	registerAllAdmissionPlugins(o.Admission.Plugins)

	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(serverConfig.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}
	sharedInformers := informers.NewSharedInformerFactory(client, serverConfig.LoopbackClientConfig.Timeout)

	// only enable admission control when running in-cluster as we require a
	// kubernetes client
	if !o.StandaloneMode {
		inClusterConfig, err := restclient.InClusterConfig()
		if err != nil {
			glog.Errorf("Failed to get kube client config: %v", err)
			return nil, err
		}
		inClusterConfig.GroupVersion = &schema.GroupVersion{}

		kubeClient, err := kubeclientset.NewForConfig(inClusterConfig)
		if err != nil {
			glog.Errorf("Failed to create clientset interface: %v", err)
			return nil, err
		}

		kubeSharedInformers := kubeinformers.NewSharedInformerFactory(kubeClient, 10*time.Minute)
		serverConfig.SharedInformerFactory = kubeSharedInformers

		serverConfig.AdmissionControl, err = buildAdmission(&o, client, sharedInformers, kubeClient, kubeSharedInformers)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize admission: %v", err)
		}
	}

	config := &apiserver.Config{
		GenericConfig:         serverConfig,
		SharedInformerFactory: sharedInformers,
	}
	return config, nil
}

// buildAdmission constructs the admission chain
func buildAdmission(s *NavigatorServerOptions,
	client clientset.Interface, sharedInformers informers.SharedInformerFactory,
	kubeClient kubeclientset.Interface, kubeSharedInformers kubeinformers.SharedInformerFactory) (admission.Interface, error) {

	admissionControlPluginNames := s.Admission.PluginNames
	glog.Infof("Admission control plugin names: %v", admissionControlPluginNames)
	var err error

	pluginInitializer := navigatorinitializer.NewPluginInitializer(client, sharedInformers, kubeClient, kubeSharedInformers)
	admissionConfigProvider, err := admission.ReadAdmissionConfiguration(admissionControlPluginNames, s.Admission.ConfigFile, api.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %v", err)
	}
	return s.Admission.Plugins.NewFromPlugins(admissionControlPluginNames, admissionConfigProvider, pluginInitializer, admissionmetrics.WithControllerMetrics)
}

func (o NavigatorServerOptions) RunNavigatorServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHook("start-navigator-server-informers", func(context genericapiserver.PostStartHookContext) error {
		config.SharedInformerFactory.Start(context.StopCh)
		return nil
	})

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
