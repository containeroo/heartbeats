package server

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/containeroo/heartbeats/internal/handler"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/middleware"
	"github.com/containeroo/heartbeats/internal/service/health"
)

// NewRouter creates a new HTTP router
func NewRouter(
	webFS fs.FS,
	routePrefix string,
	api *handler.API,
	debug bool,
) (http.Handler, error) {
	root := http.NewServeMux()

	// Handler for embedded static files
	staticContent, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return nil, fmt.Errorf("web filesystem: %w", err)
	}

	fileServer := http.FileServer(http.FS(staticContent))
	root.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	root.Handle("/", api.HomeHandler()) // no Method allowed, otherwise it crashes
	root.Handle("GET /partials/", http.StripPrefix("/partials", api.PartialHandler(api.SiteRoot)))
	root.Handle("GET /healthz", api.Healthz(health.NewService()))
	root.Handle("GET /metrics", api.Metrics())
	root.Handle("POST /-/reload", api.ReloadHandler())

	// define your API routes on a sub-mux
	root.Handle("POST /bump/{id}", api.BumpHandler())
	root.Handle("GET  /bump/{id}", api.BumpHandler())
	root.Handle("POST /bump/{id}/fail", api.FailHandler())
	root.Handle("GET  /bump/{id}/fail", api.FailHandler())

	var h http.Handler = root
	if routePrefix != "" {
		logging.SystemLogger(api.Logger, nil).Info(
			"mounted under prefix",
			"event", "routes_mounted",
			"prefix", routePrefix,
		)
		h = mountUnderPrefix(h, routePrefix)
	}

	// Optional debug logging middleware.
	if debug {
		return middleware.Chain(h, middleware.LoggingMiddleware(api.Logger), middleware.RequestIDMiddleware()), nil
	}

	return middleware.Chain(h, middleware.RequestIDMiddleware()), nil
}
