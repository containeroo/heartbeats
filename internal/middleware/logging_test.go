package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("logs method, path, status and duration", func(t *testing.T) {
		t.Parallel()

		var logs strings.Builder
		logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // simulate latency
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte("short and stout")) // nolint:errcheck
		})

		req := httptest.NewRequest(http.MethodPost, "/bump", nil)
		req.Header.Set("User-Agent", "unit-test")
		req.RemoteAddr = "1.2.3.4:5678"

		rec := httptest.NewRecorder()
		logged := LoggingMiddleware(logger)(handler)

		logged.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTeapot, rec.Code)
		assert.Contains(t, logs.String(), "method=POST")
		assert.Contains(t, logs.String(), "url_path=/bump")
		assert.Contains(t, logs.String(), "status_code=418")
		assert.Contains(t, logs.String(), "remote_addr=1.2.3.4:5678")
		assert.Contains(t, logs.String(), "user_agent=unit-test")
		assert.Contains(t, logs.String(), "duration_ms=")
		assert.Contains(t, logs.String(), `msg="HTTP request"`)
	})
}
