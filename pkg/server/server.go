package server

import (
	"context"
	"embed"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"net/http"
	"sync"
	"time"
)

// Run starts the HTTP server and handles shutdown on context cancellation.
func Run(ctx context.Context, listenAddress, version, siteRoot string, templates embed.FS, logger logger.Logger, heartbeatStore *heartbeat.Store, notificationStore *notify.Store, historyStore *history.Store) error {
	router := newRouter(logger, templates, version, siteRoot, heartbeatStore, notificationStore, historyStore)

	server := &http.Server{
		Addr:         listenAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Listening on %s", listenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Error listening and serving. %s", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("Error shutting down HTTP server. %s", err)
		}
	}()
	wg.Wait()

	return nil
}
