package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/handlers"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/middleware"
	"github.com/containeroo/heartbeats/internal/notifier"
)

// NewRouter creates a new HTTP router
func NewRouter(
	staticFS fs.FS,
	siteRoot string,
	version string,
	mgr *heartbeat.Manager,
	histStore history.Store,
	disp *notifier.Dispatcher,
	logger *slog.Logger,
	debug bool,
) http.Handler {
	root := http.NewServeMux()

	// Handler for embedded static files
	staticContent, _ := fs.Sub(staticFS, "web/static")
	fileServer := http.FileServer(http.FS(staticContent))
	root.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	root.Handle("/", handlers.HomeHandler(staticFS, version)) // no Method allowed, otherwise it crashes
	root.Handle("GET /partials/", http.StripPrefix("/partials", handlers.PartialHandler(staticFS, siteRoot, mgr, histStore, disp, logger)))
	root.Handle("GET /healthz", handlers.Healthz())
	root.Handle("GET /metrics", handlers.Metrics())

	// define your API routes on a sub-mux
	api := http.NewServeMux()
	api.Handle("POST /bump/{id}", handlers.BumpHandler(mgr, histStore, logger))
	api.Handle("GET  /bump/{id}", handlers.BumpHandler(mgr, histStore, logger))
	api.Handle("POST /bump/{id}/fail", handlers.FailHandler(mgr, histStore, logger))
	api.Handle("GET  /bump/{id}/fail", handlers.FailHandler(mgr, histStore, logger))

	// mount under /api/v1/
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	// wrap the whole mux in logging if debug
	var h http.Handler = root
	if debug {
		logMw := middleware.LoggingMiddleware(logger)
		h = logMw(h)
	}

	return h
}
