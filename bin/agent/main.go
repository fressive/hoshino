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
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"rina.icu/hoshino/agent"
)

const (
	banner = `
 __  __  ______  ______  __  __  __  __   __  ______    
/\ \_\ \/\  __ \/\  ___\/\ \_\ \/\ \/\ "-.\ \/\  __ \   
\ \  __ \ \ \/\ \ \___  \ \  __ \ \ \ \ \-.  \ \ \/\ \  
 \ \_\ \_\ \_____\/\_____\ \_\ \_\ \_\ \_\\"\_\ \_____\ 
  \/_/\/_/\/_____/\/_____/\/_/\/_/\/_/\/_/ \/_/\/_____/ 
                                                        
          ______  ______  ______  __   __  ______      
         /\  __ \/\  ___\/\  ___\/\ "-.\ \/\__  _\     
         \ \  __ \ \ \__ \ \  __\\ \ \-.  \/_/\ \/     
          \ \_\ \_\ \_____\ \_____\ \_\\"\_\ \ \_\     
           \/_/\/_/\/_____/\/_____/\/_/ \/_/  \/_/     
`
)

var (
	cmd = &cobra.Command{
		Use:   "hoshino-agent",
		Short: "The container service of Hoshino.",
		Run: func(_ *cobra.Command, _ []string) {
			instanceConfig := &agent.Config{
				Mode:           viper.GetString("mode"),
				Addr:           viper.GetString("address"),
				Port:           viper.GetInt("port"),
				DataDir:        viper.GetString("data_dir"),
				Secret:         viper.GetString("secret"),
				CertFile:       viper.GetString("cert_file"),
				PrivateKeyFile: viper.GetString("private_key_file"),
			}

			if err := instanceConfig.Validate(); err != nil {
				panic(err)
			}

			printGreetings()

			ctx, cancel := context.WithCancel(context.Background())

			go agent.RunAgentServer(instanceConfig)
			slog.Info(fmt.Sprintf("Hoshino agent server is running on %s:%d", instanceConfig.Addr, instanceConfig.Port))

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
	viper.SetConfigName("hoshino_agent")
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
