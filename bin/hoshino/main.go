// Copyright 2025 Rina
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/server"
	"rina.icu/hoshino/server/config"
	"rina.icu/hoshino/server/version"
)

const (
	banner = `
 __  __   ______   ______   __  __   __   __   __   ______   
/\ \_\ \ /\  __ \ /\  ___\ /\ \_\ \ /\ \ /\ "-.\ \ /\  __ \  
\ \  __ \\ \ \/\ \\ \___  \\ \  __ \\ \ \\ \ \-.  \\ \ \/\ \ 
 \ \_\ \_\\ \_____\\/\_____\\ \_\ \_\\ \_\\ \_\\"\_\\ \_____\
  \/_/\/_/ \/_____/ \/_____/ \/_/\/_/ \/_/ \/_/ \/_/ \/_____/

`
)

var (
	cmd = &cobra.Command{
		Use:   "hoshino",
		Short: "Hoshino is a lightweight CTF platform designed for team internal training.",
		Run: func(_ *cobra.Command, _ []string) {
			instanceConfig := &config.Config{
				Mode:       viper.GetString("mode"),
				Address:    viper.GetString("address"),
				Port:       viper.GetInt("port"),
				DataDir:    viper.GetString("data_dir"),
				DSN:        viper.GetString("dsn"),
				Driver:     viper.GetString("driver"),
				Secret:     viper.GetString("secret"),
				Kubeconfig: viper.GetString("kube_config"),
				SMTP: func() config.SMTP {
					var smtp config.SMTP
					if err := viper.UnmarshalKey("smtp", &smtp); err != nil {
						panic(err)
					}
					return smtp
				}(),
				CORS: func() config.CORS {
					var cors config.CORS
					if err := viper.UnmarshalKey("cors", &cors); err != nil {
						panic(err)
					}
					return cors
				}(),

				Version: version.Version,
			}

			if err := instanceConfig.Validate(); err != nil {
				panic(err)
			}

			printGreetings()
			slog.Info(fmt.Sprintf("Hoshino v%s", version.Version))

			ctx, cancel := context.WithCancel(context.Background())

			// Agents
			kubeconfig := flag.String("kubeconfig", instanceConfig.Kubeconfig, "(optional) absolute path to the kubeconfig file")
			flag.Parse()

			k8sConf, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
			if err != nil {
				panic(err.Error())
			}

			clientset, err := kubernetes.NewForConfig(k8sConf)
			if err != nil {
				panic(err.Error())
			}

			nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}

			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "challenge-containers",
				},
			}

			clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

			for _, node := range nodes.Items {
				slog.Info(fmt.Sprintf("Kubernetes Node Name: %s", node.Name))
			}

			// Email

			util.InitSMTP(&instanceConfig.SMTP)

			// Web

			s, err := server.NewServer(context.Background(), instanceConfig, clientset)

			if err != nil {
				cancel()
				slog.Error("Failed to create Hoshino server")
				panic(err)
			}

			if err = s.Startup(); err != nil {
				cancel()
				slog.Error("Failed to startup Hoshino server.")
				panic(err)
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-c
				cancel()
			}()
			<-ctx.Done()
		},
	}
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelInfo.Level())

	viper.SetConfigType("yaml")
	viper.SetConfigName("hoshino")
	viper.AddConfigPath(".")

	viper.SetDefault("mode", "demo")
	viper.SetDefault("address", "127.0.0.1")
	viper.SetDefault("port", 1270)
	viper.SetDefault("data_dir", "data")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func printGreetings() {
	print(banner)

	fmt.Printf(`------
Github: %s
Documentation: %s
------

`, "https://github.com/fressive/hoshino", "https://hoshino.rina.icu")

}

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
