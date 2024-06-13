package server

import (
	"heartbeats/pkg/handlers"
	"heartbeats/pkg/logger"
	"io/fs"
	"net/http"
)

// newRouter creates a new Server mux and appends Handlers
func newRouter(logger logger.Logger, staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Handler for embedded static files
	staticContent, _ := fs.Sub(staticFS, "web/static")
	fileServer := http.FileServer(http.FS(staticContent))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.Handle("GET /", handlers.Heartbeats(logger, staticFS))
	mux.Handle("GET /ping/{id}", handlers.Ping(logger))
	mux.Handle("POST /ping/{id}", handlers.Ping(logger))
	mux.Handle("GET /history/{id}", handlers.History(logger, staticFS))
	mux.Handle("GET /healthz", handlers.Healthz(logger))
	mux.Handle("POST /healthz", handlers.Healthz(logger))
	mux.Handle("GET /metrics", handlers.Metrics(logger))

	return mux
}
