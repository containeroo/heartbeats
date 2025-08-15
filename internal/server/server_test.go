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

// TestRun verifies server lifecycle: start → graceful shutdown on context cancel.
func TestRun(t *testing.T) {
	//	t.Parallel()

	t.Run("Start and stop server lifecycle", func(t *testing.T) {
		//		t.Parallel()

		var logBuf strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))

		// Auto-cancel after a short timeout to trigger shutdown.
		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()

		var wg sync.WaitGroup
		wg.Go(func() {
			// Do NOT call Done here; wg.Go handles Add/Done automatically.
			err := Run(ctx, ":0", mux, logger) // :0 = choose a random free port
			assert.NoError(t, err, "Run should not return an error")
		})

		// Give the server a moment to start before waiting for ctx timeout.
		time.Sleep(100 * time.Millisecond)

		// Wait for Run to return after ctx cancels and shutdown completes.
		wg.Wait()

		// Validate log messages.
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "starting server", "Log should contain 'starting server'")
		assert.Contains(t, logOutput, "shutting down server", "Log should contain 'shutting down server'")
	})
}
