package templates

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "template.tmpl")
		require.NoError(t, os.WriteFile(path, []byte("{{ .Value }}"), 0o600))

		tmpl, err := Load(path)
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, []byte("ok"), out)
	})

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		_, err := Load("")

		require.Error(t, err)
	})

	t.Run("error on missing file", func(t *testing.T) {
		t.Parallel()

		_, err := Load(filepath.Join(t.TempDir(), "missing.tmpl"))

		require.Error(t, err)
	})

	t.Run("error on invalid template", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "invalid.tmpl")
		require.NoError(t, os.WriteFile(path, []byte("{{ .Value "), 0o600))

		_, err := Load(path)

		require.Error(t, err)
	})
}

func TestLoadDefault(t *testing.T) {
	t.Parallel()

	tmplFS := fstest.MapFS{
		"templates/default.tmpl": {Data: []byte("{{ .Value }}")},
	}

	tmpl, err := LoadDefault(tmplFS)
	require.NoError(t, err)

	out, err := tmpl.Render(map[string]string{"Value": "ok"})
	require.NoError(t, err)
	require.Equal(t, []byte("ok"), out)
}

func TestLoadFromFS(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tmplFS := fstest.MapFS{
			"tmpl.tmpl": {Data: []byte("{{ .Value }}")},
		}

		tmpl, err := LoadFromFS(tmplFS, "tmpl.tmpl")
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, []byte("ok"), out)
	})

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		_, err := LoadFromFS(fstest.MapFS{}, "")

		require.Error(t, err)
	})

	t.Run("error on missing file", func(t *testing.T) {
		t.Parallel()

		_, err := LoadFromFS(fstest.MapFS{}, "missing.tmpl")

		require.Error(t, err)
	})

	t.Run("error on invalid template", func(t *testing.T) {
		t.Parallel()

		tmplFS := fstest.MapFS{
			"invalid.tmpl": {Data: []byte("{{ .Value ")},
		}

		_, err := LoadFromFS(tmplFS, "invalid.tmpl")

		require.Error(t, err)
	})
}

func TestParseStringTemplate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tmpl, err := ParseStringTemplate("name", "{{ .Value }}")
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, "ok", out)
	})

	t.Run("error on empty input", func(t *testing.T) {
		t.Parallel()

		_, err := ParseStringTemplate("empty", "")

		require.Error(t, err)
	})

	t.Run("error on invalid template", func(t *testing.T) {
		t.Parallel()

		_, err := ParseStringTemplate("invalid", "{{ .Value ")

		require.Error(t, err)
	})
}

func TestLoadString(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "template.tmpl")
		require.NoError(t, os.WriteFile(path, []byte("{{ .Value }}"), 0o600))

		tmpl, err := LoadString(path)
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, "ok", out)
	})

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		_, err := LoadString("")

		require.Error(t, err)
	})

	t.Run("error on missing file", func(t *testing.T) {
		t.Parallel()

		_, err := LoadString(filepath.Join(t.TempDir(), "missing.tmpl"))

		require.Error(t, err)
	})

	t.Run("error on invalid template", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "invalid.tmpl")
		require.NoError(t, os.WriteFile(path, []byte("{{ .Value "), 0o600))

		_, err := LoadString(path)

		require.Error(t, err)
	})
}

func TestLoadStringFromFS(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tmplFS := fstest.MapFS{
			"string.tmpl": {Data: []byte("{{ .Value }}")},
		}

		tmpl, err := LoadStringFromFS(tmplFS, "string.tmpl")
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, "ok", out)
	})

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		_, err := LoadStringFromFS(fstest.MapFS{}, "")

		require.Error(t, err)
	})

	t.Run("error on missing file", func(t *testing.T) {
		t.Parallel()

		_, err := LoadStringFromFS(fstest.MapFS{}, "missing.tmpl")

		require.Error(t, err)
	})

	t.Run("error on invalid template", func(t *testing.T) {
		t.Parallel()

		tmplFS := fstest.MapFS{
			"invalid.tmpl": {Data: []byte("{{ .Value ")},
		}

		_, err := LoadStringFromFS(tmplFS, "invalid.tmpl")

		require.Error(t, err)
	})
}

func TestTemplateRender(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tmplFS := fstest.MapFS{
			"tmpl.tmpl": {Data: []byte("{{ .Value }}")},
		}

		tmpl, err := LoadFromFS(tmplFS, "tmpl.tmpl")
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "payload"})
		require.NoError(t, err)
		require.Equal(t, []byte("payload"), out)
	})

	t.Run("error on missing key", func(t *testing.T) {
		t.Parallel()

		tmplFS := fstest.MapFS{
			"tmpl.tmpl": {Data: []byte("{{ .Value }}")},
		}

		tmpl, err := LoadFromFS(tmplFS, "tmpl.tmpl")
		require.NoError(t, err)

		_, err = tmpl.Render(map[string]string{})

		require.Error(t, err)
	})
}

func TestStringTemplateRender(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tmpl, err := ParseStringTemplate("string", "{{ .Value }}")
		require.NoError(t, err)

		out, err := tmpl.Render(map[string]string{"Value": "payload"})
		require.NoError(t, err)
		require.Equal(t, "payload", out)
	})

	t.Run("error on missing key", func(t *testing.T) {
		t.Parallel()

		tmpl, err := ParseStringTemplate("string", "{{ .Value }}")
		require.NoError(t, err)

		_, err = tmpl.Render(map[string]string{})

		require.Error(t, err)
	})
}

func TestFuncMapIntegration(t *testing.T) {
	t.Parallel()

	tmpl, err := ParseStringTemplate(
		"helpers",
		`{{ .Channel | default "alertmanager" | withPrefix "#" }} {{ .Title | json }} {{ .Since | formatDuration }} {{ .Tags | join "," }} {{ .LastSeen | ago }}`,
	)
	require.NoError(t, err)

	out, err := tmpl.Render(map[string]any{
		"Channel":  "",
		"Title":    `Heartbeat "missing"`,
		"Since":    90 * time.Second,
		"Tags":     []string{"prod", "api"},
		"LastSeen": time.Time{},
	})
	require.NoError(t, err)

	require.Equal(t, `#alertmanager "Heartbeat \"missing\"" 1m30s prod,api never`, out)
}

func TestJoin(t *testing.T) {
	t.Parallel()

	require.Equal(t, "prod,api", join(",", []string{"prod", "api"}))
	require.Equal(t, "prod, api", join(", ", []string{"prod", "api"}))
	require.Equal(t, "prod", join(",", []string{"prod"}))
	require.Equal(t, "", join(",", nil))
	require.Equal(t, "", join(",", []string{}))
}
