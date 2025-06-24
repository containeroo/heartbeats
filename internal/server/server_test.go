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
	//	t.Parallel()

	t.Run("Start and stop server lifecycle", func(t *testing.T) {
		//		t.Parallel()

		var logBuf strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))

		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Run(ctx, ":0", mux, logger) // :0 = random port
			assert.NoError(t, err, "Run function should not return an error")
		}()

		time.Sleep(100 * time.Millisecond) // Give the server time to start
		wg.Wait()

		// Validate log messages
		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "starting server", "Log should contain 'starting server'")
		assert.Contains(t, logOutput, "shutting down server", "Log should contain 'shutting down server'")
	})
}
