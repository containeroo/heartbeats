package main

import (
	"context"
	"embed"
	"fmt"
	"heartbeats/pkg/config"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/server"
	"os"

	"github.com/alecthomas/kingpin/v2"
)

const version = "0.6.8"

//go:embed web
var templates embed.FS

var (
	configPath    = kingpin.Flag("config", "Path to the configuration file").Short('c').Envar("HEARTBEATS_CONFIG").Default("./deploy/config.yaml").String()
	listenAddress = kingpin.Flag("listen-address", "Address to listen on").Short('l').Envar("HEARTBEATS_LISTEN_ADDRESS").Default("localhost:8080").String()
	siteRoot      = kingpin.Flag("site-root", "Site root for the heartbeat service").Short('s').Envar("HEARTBEATS_SITE_ROOT").Default("http://<listenaddress>").String()
	maxSize       = kingpin.Flag("max-size", "Maximum size of the cache").Short('m').Envar("HEARTBEATS_MAX_SIZE").Default("1000").Int()
	reduce        = kingpin.Flag("reduce", "Percentage to reduce when max size is exceeded").Short('r').Envar("HEARTBEATS_REDUCE").Default("25").Int()
	verbose       = kingpin.Flag("verbose", "Enable verbose logging").Short('v').Envar("HEARTBEATS_VERBOSE").Bool()
)

func run(ctx context.Context, verbose bool) error {
	kingpin.UsageTemplate(CompactUsageTemplate)
	kingpin.Version(version)
	kingpin.Parse()

	log := logger.NewLogger(verbose)

	heartbeatsStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()
	historyStore := history.NewStore()

	if err := config.Read(
		*configPath,
		heartbeatsStore,
		notificationStore,
		historyStore,
		*maxSize,
		*reduce,
	); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	return server.Run(
		ctx,
		*listenAddress,
		version,
		*siteRoot,
		templates,
		log,
		heartbeatsStore,
		notificationStore,
		historyStore,
	)
}

func main() {
	ctx := context.Background()
	if err := run(ctx, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

var CompactUsageTemplate = `{{define "FormatCommand" -}}
{{if .FlagSummary}} {{.FlagSummary}}{{end -}}
{{range .Args}}{{if not .Hidden}} {{if not .Required}}[{{end}}{{if .PlaceHolder}}{{.PlaceHolder}}{{else}}<{{.Name}}>{{end}}{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}{{end -}}
{{end -}}

{{define "FormatCommandList" -}}
{{range . -}}
{{if not .Hidden -}}
{{.Depth|Indent}}{{.Name}}{{if .Default}}*{{end}}{{template "FormatCommand" .}}
{{end -}}
{{template "FormatCommandList" .Commands -}}
{{end -}}
{{end -}}

{{define "FormatUsage" -}}
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0 -}}
{{end -}}

{{end -}}

{{if .Context.SelectedCommand -}}
usage: {{.App.Name}} {{.Context.SelectedCommand}}{{template "FormatUsage" .Context.SelectedCommand}}
{{else -}}
usage: {{.App.Name}}{{template "FormatUsage" .App}}
{{end -}}
{{if .Context.Flags -}}
Flags:
{{range .Context.Flags -}}
  {{if .Short}}-{{.Short}}, {{end}}--{{.Name}} {{.Help}} {{if .Default}}(default: {{.Default}}){{end}} {{if .Envar}}[env: {{.Envar}}]{{end}}
{{end -}}
{{end -}}
{{if .Context.Args -}}
Args:
{{range .Context.Args -}}
  --{{.Name}} {{.Help}}
{{end -}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{if .Context.SelectedCommand.Commands -}}
Commands:
  {{.Context.SelectedCommand}}
{{template "FormatCommandList" .Context.SelectedCommand.Commands}}
{{end -}}
{{else if .App.Commands -}}
Commands:
{{template "FormatCommandList" .App.Commands}}
{{end -}}
`
