package debugserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/handlers"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/notifier"
)

// Run starts the local-only debug server for manual testing.
func Run(ctx context.Context, port int, mgr *heartbeat.Manager, dispatcher *notifier.Dispatcher, logger *slog.Logger) {
	mux := http.NewServeMux()

	mux.Handle("GET /internal/receiver/{id}", handlers.TestReceiverHandler(dispatcher, logger))
	mux.Handle("GET /internal/heartbeat/{id}", handlers.TestHeartbeatHandler(mgr, logger))

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Info("starting debug server", "listenAddr", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown on ctx cancel.
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	// Serve requests on 127.0.0.1 until shutdown.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("debug server error", "error", err)
		}
	}()
}
