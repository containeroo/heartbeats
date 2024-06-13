package main

import (
	"context"
	"embed"
	"fmt"
	"heartbeats/pkg/config"
	"heartbeats/pkg/flags"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/server"
	"os"
)

const version = "0.6.8"

//go:embed web
var templates embed.FS

func run(ctx context.Context, verbose bool) error {
	config.App.Version = version

	result := flags.ParseFlags(os.Args, os.Stdout)

	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", result.Err)
		os.Exit(1)
	}

	if result.ShowHelp {
		os.Exit(0)
	}

	if result.ShowVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	log := logger.NewLogger(verbose)

	if err := config.App.Read(); err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	return server.Run(ctx, config.App.Server.ListenAddress, templates, log)
}

func main() {
	ctx := context.Background()
	if err := run(ctx, config.App.Verbose); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
