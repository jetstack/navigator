// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/jetstack-experimental/navigator/cmd/app"
	intclient "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"
	intinformers "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/externalversions"
	"github.com/jetstack-experimental/navigator/pkg/kube"
	"github.com/jetstack-experimental/navigator/pkg/tpr"
)

var cfgFile string
var apiServerHost string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "navigator",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	// TODO: Refactor this function from this package
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := kube.Config(apiServerHost)

		if err != nil {
			logrus.Fatalf("error creating rest config: %s", err.Error())
		}

		intcl, err := intclient.NewForConfig(cfg)

		if err != nil {
			logrus.Fatalf("error creating internal group client: %s", err.Error())
		}

		cl, err := kubernetes.NewForConfig(cfg)

		if err != nil {
			logrus.Fatalf("error creating kubernetes client: %s", err.Error())
		}

		if err = tpr.Ensure(cl); err != nil {
			logrus.Fatalf("error creating ThirdPartyResources: %s", err.Error())
		}

		tprClient, err := kube.NewMarshalRESTClient(apiServerHost)

		if err != nil {
			logrus.Fatalf("error creating third party resource client: %s", err.Error())
		}

		ctx := app.ControllerContext{
			Client:                 cl,
			TPRClient:              tprClient,
			InformerFactory:        informers.NewSharedInformerFactory(cl, time.Second*30),
			MarshalInformerFactory: intinformers.NewSharedInformerFactory(intcl, time.Second*30),
			Namespace:              metav1.NamespaceAll,
			Stop:                   make(<-chan struct{}),
		}

		err = app.StartControllers(
			&ctx,
			app.Known(),
			ctx.Stop,
		)

		if err != nil {
			logrus.Fatalf("error running controllers: %s", err.Error())
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.navigator.yaml)")
	RootCmd.PersistentFlags().StringVar(&apiServerHost, "apiServerHost", "", "optional api server hostname override")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".navigator") // name of config file (without extension)
	viper.AddConfigPath("$HOME")      // adding home directory as first search path
	viper.AutomaticEnv()              // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
