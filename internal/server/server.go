package server

import (
	"context"
	"embed"
	"heartbeats/internal/logger"
	"net/http"
	"sync"
	"time"
)

var StaticFS embed.FS

func Run(ctx context.Context, listenAddress string, templates embed.FS, logger logger.Logger) error {
	router := newRouter(logger, templates)

	server := &http.Server{
		Addr:         listenAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	go func() {
		logger.Infof("listening on %s", listenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("error listening and serving: %s", err)
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
			logger.Errorf("error shutting down http server: %s", err)
		}
	}()
	wg.Wait()

	return nil
}
