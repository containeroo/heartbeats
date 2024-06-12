package server

import (
	"embed"
	"heartbeats/pkg/handlers"
	"heartbeats/pkg/logger"
	"io/fs"
	"net/http"
)

// newRouter creates a new HTTP server mux, setting up routes and handlers
func newRouter(logger logger.Logger, staticFS embed.FS) http.Handler {
	mux := http.NewServeMux()

	// handler for embed static files
	fsys := fs.FS(staticFS)
	contentStatic, _ := fs.Sub(fsys, "web/static")
	fs := http.FileServer(http.FS(contentStatic))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.Handle("GET /", handlers.Heartbeats(logger, staticFS))

	mux.Handle("GET /ping/{id}", handlers.Ping(logger))
	mux.Handle("POST /ping/{id}", handlers.Ping(logger))
	mux.Handle("GET /history/{id}", handlers.History(logger, staticFS))
	mux.Handle("GET /healthz", handlers.Healthz(logger))
	mux.Handle("POST /healthz", handlers.Healthz(logger))
	mux.Handle("GET /metrics", handlers.Metrics(logger))

	return mux
}
