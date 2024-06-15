package server

import (
	"heartbeats/pkg/handlers"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"io/fs"
	"net/http"
)

// newRouter creates a new Server mux and appends Handlers
func newRouter(logger logger.Logger, staticFS fs.FS, version, siteRoot string, heartbeatStore *heartbeat.Store, notificationStore *notify.Store, historyStore *history.Store) http.Handler {
	mux := http.NewServeMux()

	// Handler for embedded static files
	staticContent, _ := fs.Sub(staticFS, "web/static")
	fileServer := http.FileServer(http.FS(staticContent))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.Handle("GET /", handlers.Heartbeats(logger, staticFS, version, siteRoot, heartbeatStore, notificationStore))
	mux.Handle("GET /ping/{id}", handlers.Ping(logger, heartbeatStore, notificationStore, historyStore))
	mux.Handle("POST /ping/{id}", handlers.Ping(logger, heartbeatStore, notificationStore, historyStore))
	mux.Handle("GET /history/{id}", handlers.History(logger, staticFS, version, heartbeatStore, historyStore))
	mux.Handle("GET /healthz", handlers.Healthz(logger))
	mux.Handle("POST /healthz", handlers.Healthz(logger))
	mux.Handle("GET /metrics", handlers.Metrics(logger))

	return mux
}
