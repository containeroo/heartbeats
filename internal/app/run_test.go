package app_test

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/containeroo/heartbeats/internal/app"
	"github.com/stretchr/testify/assert"
)

// waitForLog polls the log buffer until it contains the expected substring or times out.
func waitForLog(t *testing.T, buf *bytes.Buffer, contains string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), contains) {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}

	return false
}

func TestRun(t *testing.T) {
	t.Parallel()

	// in-memory file system with minimal template files
	webFS := fstest.MapFS{
		"web/static/css/heartbeats.css": &fstest.MapFile{Data: []byte(`body {}`)},
		"web/templates/base.html":       &fstest.MapFile{Data: []byte(`{{define "base"}}<html>{{template "navbar"}}<footer>{{.Version}}</footer>{{end}}`)},
		"web/templates/navbar.html":     &fstest.MapFile{Data: []byte(`{{define "navbar"}}<nav>nav</nav>{{end}}`)},
		"web/templates/heartbeats.html": &fstest.MapFile{Data: []byte(`heartbeat page`)},
		"web/templates/receivers.html":  &fstest.MapFile{Data: []byte(`receiver page`)},
		"web/templates/history.html":    &fstest.MapFile{Data: []byte(`history page`)},
		"web/templates/footer.html":     &fstest.MapFile{Data: []byte(`{{define "footer"}}<!-- footer -->{{end}}`)},
	}

	t.Run("shows help and exits", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "dev", "abc", []string{"--help"}, &out, os.Getenv)

		assert.NoError(t, err)
		assert.Contains(t, out.String(), "Usage:")
	})

	t.Run("shows version and exits", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "v1.2.3", "deadbeef", []string{"--version"}, &out, os.Getenv)

		assert.NoError(t, err)
		assert.Equal(t, "Heartbeats version v1.2.3\n", out.String())
	})

	t.Run("fails with invalid log format", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "dev", "abc", []string{"--log-format", "xml"}, &out, os.Getenv)

		assert.EqualError(t, err, "invalid CLI flags: invalid log format: 'xml'")
	})

	t.Run("fails with invalid retry delay format", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "dev", "abc", []string{"--retry-delay", "foo"}, &out, os.Getenv)

		assert.EqualError(t, err, "parsing error: invalid argument \"foo\" for \"--retry-delay\" flag: time: invalid duration \"foo\"")
	})

	t.Run("fails with invalid retry count", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "dev", "abc", []string{"--retry-count", "0"}, &out, os.Getenv)

		assert.EqualError(t, err, "invalid CLI flags: retry count must be -1 (infinite) or >= 1, got 0")
	})

	t.Run("fails with invalid retry delay", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "dev", "abc", []string{"--retry-delay", "200ms"}, &out, os.Getenv)

		assert.EqualError(t, err, "invalid CLI flags: retry delay must be at least 1s, got 200ms")
	})

	t.Run("fails when config file is missing", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		err := app.Run(context.Background(), webFS, "dev", "abc", []string{"--config", "nope.yaml"}, &out, os.Getenv)

		assert.EqualError(t, err, "failed to load config: open nope.yaml: no such file or directory")
	})

	t.Run("fails when YAML config is invalid", func(t *testing.T) {
		t.Parallel()

		tmpFile := filepath.Join(t.TempDir(), "bad.yaml")
		assert.NoError(t, os.WriteFile(tmpFile, []byte("receivers:"), 0644))

		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
		defer cancel()

		var out bytes.Buffer
		err := app.Run(ctx, webFS, "dev", "abc", []string{"--config", tmpFile}, &out, os.Getenv)

		assert.EqualError(t, err, "invalid YAML config: at least one heartbeat must be defined")
	})

	t.Run("fails on history init", func(t *testing.T) {
		webFS := webFS
		tmpFile := filepath.Join(t.TempDir(), "good.yaml")

		config := `
receivers:
  team1:
    slack_configs:
      - channel: "#alerts"
        token: "dummy"
heartbeats:
  ping:
    interval: 500ms
    grace: 500ms
    receivers: ["team1"]
`
		assert.NoError(t, os.WriteFile(tmpFile, []byte(config), 0644))

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		var buf bytes.Buffer
		args := []string{
			"--config", tmpFile,
			"--listen-address", ":8070",
			"--debug",
			"--history-backend", "badger",
			"--badger-path", "",
			"--retry-count", "2",
			"--retry-delay", "1s",
		}
		err := app.Run(ctx, webFS, "dev", "abc", args, &buf, os.Getenv)
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to initialize history: badger backend requires a path")
	})

	t.Run("startup and state change succeeds", func(t *testing.T) {
		t.Parallel()

		webFS := webFS
		tmpFile := filepath.Join(t.TempDir(), "good.yaml")

		config := `
receivers:
  team1:
    slack_configs:
      - channel: "#alerts"
        token: "dummy"
heartbeats:
  ping:
    interval: 500ms
    grace: 500ms
    receivers: ["team1"]
`
		assert.NoError(t, os.WriteFile(tmpFile, []byte(config), 0644))

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		var buf bytes.Buffer
		args := []string{
			"--config", tmpFile,
			"--listen-address", ":8070",
			"--debug",
			"--retry-count", "2",
			"--retry-delay", "1s",
		}

		go func() {
			_ = app.Run(ctx, webFS, "dev", "abc", args, &buf, os.Getenv)
		}()

		time.Sleep(200 * time.Millisecond) // wait to start the server

		resp, err := http.Post("http://localhost:8070/bump/ping", "text/plain", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.True(t, waitForLog(t, &buf, `"from":"active","to":"grace"`, 2*time.Second), "expected Active → Grace")
		assert.True(t, waitForLog(t, &buf, `"from":"grace","to":"missing"`, 2*time.Second), "expected Grace → Missing")
		assert.True(t, waitForLog(t, &buf, `"msg":"retrying","attempt":1`, 2*time.Second), "expected retry 1")
		assert.True(t, waitForLog(t, &buf, `"msg":"retrying","attempt":2`, 2*time.Second), "expected retry 2")
		assert.True(t, waitForLog(t, &buf, `notification failed after 2 retries`, 2*time.Second), "expected failure log")
	})
}
