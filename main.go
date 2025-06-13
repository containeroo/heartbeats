package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/containeroo/heartbeats/internal/app"
)

var (
	Version string = "dev"
	Commit  string = "none"
)

//go:embed web
var webFS embed.FS

// main sets up the application context and runs the proxy.
func main() {
	ctx := context.Background()

	if err := app.Run(ctx, webFS, Version, Commit, os.Args[1:], os.Stdout, os.Getenv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
