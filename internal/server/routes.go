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
	webFS fs.FS,
	siteRoot string,
	version string,
	mgr *heartbeat.Manager,
	hist history.Store,
	disp *notifier.Dispatcher,
	logger *slog.Logger,
	debug bool,
) http.Handler {
	root := http.NewServeMux()

	// Handler for embedded static files
	staticContent, _ := fs.Sub(webFS, "web/static")
	fileServer := http.FileServer(http.FS(staticContent))
	root.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	root.Handle("/", handlers.HomeHandler(webFS, version)) // no Method allowed, otherwise it crashes
	root.Handle("GET /partials/", http.StripPrefix("/partials", handlers.PartialHandler(webFS, siteRoot, mgr, hist, disp, logger)))
	root.Handle("GET /healthz", handlers.Healthz())
	root.Handle("GET /metrics", handlers.Metrics(hist))

	// define your API routes on a sub-mux
	root.Handle("POST /bump/{id}", handlers.BumpHandler(mgr, hist, logger))
	root.Handle("GET  /bump/{id}", handlers.BumpHandler(mgr, hist, logger))
	root.Handle("POST /bump/{id}/fail", handlers.FailHandler(mgr, hist, logger))
	root.Handle("GET  /bump/{id}/fail", handlers.FailHandler(mgr, hist, logger))

	// wrap the whole mux in logging if debug
	if debug {
		return middleware.Chain(root, middleware.LoggingMiddleware(logger))
	}

	return root
}
