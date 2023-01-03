/*
Copyright © 2022 containeroo github.com/containeroo/heartbeats

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
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/containeroo/heartbeats/internal"
	"github.com/containeroo/heartbeats/internal/cache"
	"github.com/containeroo/heartbeats/internal/docs"
	"github.com/containeroo/heartbeats/internal/server"
	"github.com/fsnotify/fsnotify"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	version = "v0.4.4"
)

var debug, trace, jsonLog bool
var printVersion bool
var StaticFs embed.FS

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s\n", entry.Time.Format(time.RFC3339), entry.Message)), nil
}
func toggleDebug(cmd *cobra.Command, args []string) {

	if jsonLog {
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
	Long: `Small helper service to monitor heartbeats (repeating "pings" from other systems).
If a "ping" does not arrive in the given interval & grace period, Heartbeats will send notifications.`,

	PersistentPreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		if printVersion {
			fmt.Println(version)
			os.Exit(0)
		}
		internal.HeartbeatsServer.Version = version

		if internal.HeartbeatsServer.Server.SiteRoot == "" {
			internal.HeartbeatsServer.Server.SiteRoot = fmt.Sprintf("http://%s:%d", internal.HeartbeatsServer.Server.Hostname, internal.HeartbeatsServer.Server.Port)
		}

		if err := internal.ReadConfigFile(internal.HeartbeatsServer.Config.Path, true); err != nil {
			log.Fatal(err)
		}

		// Initialize the cache
		cache.Local = cache.New(internal.HeartbeatsServer.Cache.MaxSize, internal.HeartbeatsServer.Cache.Reduce)

		// Initialize the documentation
		c := docs.Cache{
			MaxSize: internal.HeartbeatsServer.Cache.MaxSize,
			Reduce:  internal.HeartbeatsServer.Cache.Reduce,
		}
		docs.Documentation = *docs.NewDocumentation(internal.HeartbeatsServer.Server.SiteRoot, &c)

		// watch config
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Infof("«%s» has changed. reload it", e.Name)

			if err := internal.ReadConfigFile(internal.HeartbeatsServer.Config.Path, false); err != nil {
				log.Fatal(err)
			}

			// regenerate documentation
			c := docs.Cache{
				MaxSize: internal.HeartbeatsServer.Cache.MaxSize,
				Reduce:  internal.HeartbeatsServer.Cache.Reduce,
			}
			docs.Documentation = *docs.NewDocumentation(internal.HeartbeatsServer.Server.SiteRoot, &c)
		})
		viper.WatchConfig()

		server.RunServer(internal.HeartbeatsServer.Server.Hostname, internal.HeartbeatsServer.Server.Port)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().StringVarP(&internal.HeartbeatsServer.Config.Path, "config", "c", "./config.yaml", "Path to Heartbeats config file")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Verbose logging.")
	rootCmd.Flags().BoolVarP(&trace, "trace", "t", false, "More verbose logging.")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "trace")
	rootCmd.Flags().BoolVarP(&jsonLog, "json-log", "j", false, "Output logging as json.")

	rootCmd.Flags().BoolVarP(&printVersion, "version", "v", false, "Print the current version and exit.")
	rootCmd.Flags().StringVar(&internal.HeartbeatsServer.Server.Hostname, "host", "127.0.0.1", "Host of Heartbeat service.")
	rootCmd.Flags().IntVarP(&internal.HeartbeatsServer.Server.Port, "port", "p", 8090, "Port to listen on")
	rootCmd.Flags().StringVarP(&internal.HeartbeatsServer.Server.SiteRoot, "site-root", "s", "", "Site root for the heartbeat service (default \"http://host:port\")")

	rootCmd.Flags().IntVarP(&internal.HeartbeatsServer.Cache.MaxSize, "max-size", "m", 500, "Max Size of History Cache per Heartbeat")
	rootCmd.Flags().IntVarP(&internal.HeartbeatsServer.Cache.Reduce, "reduce", "r", 100, "Reduce Max Size of History Cache by this value if it exceeds the Max Size")

	rootCmd.Flags().SortFlags = false
}
