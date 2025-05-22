package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("Start and stop server lifecycle", func(t *testing.T) {
		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		ctx, cancel := context.WithCancel(context.Background())

		mux := http.NewServeMux()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Run(ctx, ":8080", mux, logger)
			assert.NoError(t, err, "Run function should not return an error")
		}()

		time.Sleep(100 * time.Millisecond) // Give the server time to start

		cancel() // Cancel the context to stop the server

		wg.Wait()

		// Validate log messages
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "starting server", "Log should contain 'starting server'")
		assert.Contains(t, logOutput, "shutting down server", "Log should contain 'shutting down server'")
	})
}
