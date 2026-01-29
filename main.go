package main

import (
	"context"
	"embed"
	"os"

	"github.com/containeroo/heartbeats/internal/app"
)

var (
	// Version is the build version set via ldflags.
	Version string = "dev"
	// Commit is the build commit set via ldflags.
	Commit string = "none"
)

//go:embed templates web/dist
var templatesFS embed.FS

// main sets up the application context and runs the main loop.
func main() {
	ctx := context.Background()

	if err := app.Run(ctx, templatesFS, Version, Commit, os.Args[1:], os.Stdout); err != nil {
		os.Exit(1)
	}
}
