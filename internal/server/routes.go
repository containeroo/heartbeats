package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/handlers"
	"github.com/containeroo/heartbeats/internal/middleware"
	"github.com/containeroo/heartbeats/internal/service/health"
)

// NewRouter creates a new HTTP router
func NewRouter(
	webFS fs.FS,
	siteRoot string,
	routePrefix string,
	version string,
	api *handlers.API,
	logger *slog.Logger,
	debug bool,
) http.Handler {
	root := http.NewServeMux()

	// Handler for embedded static files
	staticContent, _ := fs.Sub(webFS, "web/static")
	fileServer := http.FileServer(http.FS(staticContent))
	root.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	root.Handle("/", api.HomeHandler()) // no Method allowed, otherwise it crashes
	root.Handle("GET /partials/", http.StripPrefix("/partials", api.PartialHandler(siteRoot)))
	root.Handle("GET /healthz", api.Healthz(health.NewService()))
	root.Handle("GET /metrics", api.Metrics())

	// define your API routes on a sub-mux
	root.Handle("POST /bump/{id}", api.BumpHandler())
	root.Handle("GET  /bump/{id}", api.BumpHandler())
	root.Handle("POST /bump/{id}/fail", api.FailHandler())
	root.Handle("GET  /bump/{id}/fail", api.FailHandler())

	// Mount the whole app under the prefix if provided
	var handler http.Handler = root
	if routePrefix != "" {
		handler = mountUnderPrefix(root, routePrefix)
	}

	// wrap the whole mux in logging if debug
	if debug {
		return middleware.Chain(handler, middleware.LoggingMiddleware(logger))
	}

	return handler
}
