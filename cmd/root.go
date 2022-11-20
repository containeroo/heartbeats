/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gi8lino/heartbeats/internal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	version = "v0.1.0"
)

var debug, trace, JsonLog bool

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s\n", entry.Time.Format(time.RFC3339), entry.Message)), nil
}
func toggleDebug(cmd *cobra.Command, args []string) {

	if JsonLog {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&PlainFormatter{})
	}

	if debug {
		internal.HeartbeatsServer.Config.Logging = "debug"
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logs enabled")
	}
	if trace {
		internal.HeartbeatsServer.Config.Logging = "trace"
		log.SetLevel(log.TraceLevel)
		log.Debug("Trace logs enabled")
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Wait for heartbeats and notify if they are missing",
	Long: `Heartbeats waits for heartbeats and notifies if they are missing.
You can configure the interval and grace period for each heartbeat separately and it will notify you if a heartbeat is missing.
`,

	PersistentPreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		if internal.HeartbeatsServer.Config.PrintVersion {
			fmt.Println(version)
			os.Exit(0)
		}
		internal.HeartbeatsServer.Version = version

		if internal.HeartbeatsServer.Server.SiteRoot == "" {
			internal.HeartbeatsServer.Server.SiteRoot = fmt.Sprintf("http://%s:%d", internal.HeartbeatsServer.Server.Hostname, internal.HeartbeatsServer.Server.Port)
		}

		if err := internal.ReadConfigFile(internal.HeartbeatsServer.Config.Path); err != nil {
			log.Fatal(err)
		}

		// watch config
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Infof("«%s» has changed. reload it", e.Name)
			if err := internal.ReadConfigFile(internal.HeartbeatsServer.Config.Path); err != nil {
				log.Fatal(err)
			}
		})
		viper.WatchConfig()

		// Run the server
		internal.HeartbeatsServer.Run()
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().StringVarP(&internal.HeartbeatsServer.Config.Path, "config", "c", "./config.yaml", "Path to notifications config file")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Verbose logging.")
	rootCmd.Flags().BoolVarP(&trace, "trace", "t", false, "More verbose logging.")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "trace")
	rootCmd.Flags().BoolVarP(&JsonLog, "json-log", "j", false, "Output logging as json.")

	rootCmd.Flags().BoolVarP(&internal.HeartbeatsServer.Config.PrintVersion, "version", "v", false, "Print the current version and exit.")
	rootCmd.Flags().StringVar(&internal.HeartbeatsServer.Server.Hostname, "host", "127.0.0.1", "Host of Heartbeat service.")
	rootCmd.Flags().IntVarP(&internal.HeartbeatsServer.Server.Port, "port", "p", 8090, "Port to listen on")
	rootCmd.Flags().StringVarP(&internal.HeartbeatsServer.Server.SiteRoot, "site-root", "s", "", "Site root for the heartbeat service (default \"http://host:port\")")
	rootCmd.Flags().SortFlags = false
}
