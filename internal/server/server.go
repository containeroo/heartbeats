package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Run sets up and manages the reverse proxy HTTP server.
func Run(ctx context.Context, listenAddr string, router http.Handler, logger *slog.Logger) error {
	// Create server with sensible timeouts.
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start the server in the background.
	go func() {
		logger.Info("starting server", "listenAddr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	// Graceful shutdown once the context is canceled.
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done() // wait for cancel/timeout

		logger.Info("shutting down server")

		// Use a bounded timeout to finish in-flight requests.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "err", err)
		}
	})

	// Block until the shutdown goroutine finishes.
	wg.Wait()
	return nil
}
