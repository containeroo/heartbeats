package routes

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/handler"
	"github.com/containeroo/heartbeats/internal/logging"
)

// NewRouter builds the HTTP router and applies optional route prefixing.
func NewRouter(
	appFS fs.FS,
	api *handler.API,
	routePrefix string,
	logger *slog.Logger,
) (http.Handler, error) {
	mux := http.NewServeMux()

	spaContent, err := fs.Sub(appFS, "web/dist")
	if err != nil {
		return nil, fmt.Errorf("spa filesystem: %w", err)
	}
	spaFiles := http.FileServer(http.FS(spaContent))
	mux.Handle("GET /assets/", spaFiles)
	mux.Handle("GET /fonts/", spaFiles)
	mux.Handle("GET /sounds/", spaFiles)
	mux.Handle("GET /index.html", spaFiles)
	mux.Handle("GET /favicon-16x16.png", spaFiles)
	mux.Handle("GET /heartbeats.svg", spaFiles)
	mux.Handle("GET /heartbeats-red.svg", spaFiles)
	mux.Handle("GET /favicon-32x32.png", spaFiles)
	mux.Handle("GET /favicon-48x48.png", spaFiles)
	mux.Handle("GET /favicon-64x64.png", spaFiles)
	mux.Handle("GET /apple-touch-icon.png", spaFiles)

	mux.HandleFunc("GET /healthz", api.Healthz())
	mux.HandleFunc("POST /healthz", api.Healthz())
	mux.Handle("POST /-/reload", api.ReloadHandler())
	mux.Handle("GET /metrics", api.Metrics())
	mux.Handle("POST /metrics", api.Metrics())

	apiMux := http.NewServeMux()
	apiMux.Handle("GET /config", api.Config())
	apiMux.HandleFunc("GET /heartbeats", api.HeartbeatsSum())
	apiMux.HandleFunc("GET /receivers", api.ReceiversSum())
	apiMux.HandleFunc("GET /status", api.StatusAll())
	apiMux.HandleFunc("GET /status/{id}", api.Status())
	apiMux.HandleFunc("GET /history", api.HistoryAll())
	apiMux.HandleFunc("GET /history/{id}", api.HistoryByHeartbeat())
	apiMux.HandleFunc("GET /heartbeat/{id}", HistoryMiddleware(api.HistoryRecorder(), api.Heartbeat()))
	apiMux.HandleFunc("POST /heartbeat/{id}", HistoryMiddleware(api.HistoryRecorder(), api.Heartbeat()))
	apiMux.HandleFunc("GET /ws", api.WS())

	// Mount API under /api
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// SPA
	mux.Handle("/", handler.SPA(spaContent, routePrefix))

	var h http.Handler = mux
	if routePrefix != "" {
		logger.Info("mounted under prefix",
			"event", logging.EventRoutesMounted.String(),
			"prefix", routePrefix,
		)
		h = mountUnderPrefix(h, routePrefix)
	}

	h = LoggingMiddleware(api.Logger, h)

	return h, nil
}
