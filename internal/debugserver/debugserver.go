package debugserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containeroo/heartbeats/internal/handler"
)

// Run starts the local-only debug server for manual testing.
func Run(ctx context.Context, port int, api *handler.API) {
	mux := http.NewServeMux()

	mux.Handle("GET /internal/receiver/{id}", api.TestReceiverHandler())
	mux.Handle("GET /internal/heartbeat/{id}", api.TestHeartbeatHandler())

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	api.Logger.Info("starting debug server", "listenAddr", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Direct shutdown on ctx cancel.
	go func() {
		<-ctx.Done()
		server.Close() // nolint:errcheck
	}()

	// Serve requests on 127.0.0.1 until shutdown.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			api.Logger.Error("debug server error", "error", err)
		}
	}()
}
