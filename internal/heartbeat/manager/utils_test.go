package manager

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/templates"
	"github.com/stretchr/testify/require"
)

func TestBuildReceivers(t *testing.T) {
	t.Parallel()
	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	defaultWebhook, err := templates.LoadFromFS(templateFS, "templates/default.tmpl")
	require.NoError(t, err)
	defaultTitle, err := templates.ParseStringTemplate("title", "{{ .Title }}")
	require.NoError(t, err)
	defaultEmail, err := templates.ParseStringTemplate("email", "{{ .Title }}")
	require.NoError(t, err)
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))

	cfg := &config.Config{
		Receivers: map[string]config.ReceiverConfig{
			"ops": {
				Webhooks: []config.WebhookConfig{
					{URL: "https://example.com"},
				},
				Emails: []config.EmailConfig{
					{
						Host: "smtp",
						Port: 587,
						From: "a@example.com",
						To:   []string{"b@example.com"},
					},
				},
			},
		},
	}
	hbConfig := config.HeartbeatConfig{
		Receivers: []string{"ops"},
	}

	t.Run("success builded receivers", func(t *testing.T) {
		t.Parallel()
		receivers, names, err := buildReceivers(cfg, hbConfig, defaultWebhook, defaultTitle, defaultEmail, templateFS, logger)
		require.NoError(t, err)
		require.Len(t, receivers, 1)
		require.ElementsMatch(t, []string{"ops"}, names)
	})

	t.Run("unknown receiver", func(t *testing.T) {
		t.Parallel()
		hbConfig.Receivers = []string{"missing"}
		_, _, err := buildReceivers(cfg, hbConfig, defaultWebhook, defaultTitle, defaultEmail, templateFS, logger)
		require.Error(t, err)
	})
}

func TestResolveTitleTemplate(t *testing.T) {
	t.Parallel()
	fallback, err := templates.ParseStringTemplate("fallback", "fallback")
	require.NoError(t, err)
	t.Run("receiver template wins", func(t *testing.T) {
		t.Parallel()
		out, err := resolveTitleTemplate("heartbeat", "receiver", fallback)
		require.NoError(t, err)
		res, err := out.Render(nil)
		require.NoError(t, err)
		require.Equal(t, "receiver", res)
	})
	t.Run("heartbeat template used", func(t *testing.T) {
		t.Parallel()
		out, err := resolveTitleTemplate("heartbeat", "", fallback)
		require.NoError(t, err)
		res, err := out.Render(nil)
		require.NoError(t, err)
		require.Equal(t, "heartbeat", res)
	})
	t.Run("fallback returned", func(t *testing.T) {
		t.Parallel()
		out, err := resolveTitleTemplate("", "", fallback)
		require.NoError(t, err)
		require.Equal(t, fallback, out)
	})
}

func TestResolveWebhookTemplate(t *testing.T) {
	t.Parallel()
	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	fallback, err := templates.LoadFromFS(templateFS, "templates/default.tmpl")
	require.NoError(t, err)
	t.Run("builtin loads", func(t *testing.T) {
		t.Parallel()
		out, err := resolveWebhookTemplate("slack", "", fallback, templateFS)
		require.NoError(t, err)
		require.NotNil(t, out)
	})
	t.Run("override loads", func(t *testing.T) {
		t.Parallel()
		temp := t.TempDir()
		path := temp + "/custom.tmpl"
		require.NoError(t, os.WriteFile(path, []byte("value"), 0o600))
		out, err := resolveWebhookTemplate("", path, fallback, templateFS)
		require.NoError(t, err)
		require.NotNil(t, out)
	})
}

func TestResolveEmailTemplate(t *testing.T) {
	t.Parallel()
	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	fallback, err := templates.ParseStringTemplate("fallback", "fallback")
	require.NoError(t, err)
	t.Run("builtin loads", func(t *testing.T) {
		t.Parallel()
		out, err := resolveEmailTemplate("email", "", fallback, templateFS)
		require.NoError(t, err)
		require.NotNil(t, out)
	})
	t.Run("empty uses fallback", func(t *testing.T) {
		t.Parallel()
		out, err := resolveEmailTemplate("", "", fallback, templateFS)
		require.NoError(t, err)
		require.Equal(t, fallback, out)
	})
}
