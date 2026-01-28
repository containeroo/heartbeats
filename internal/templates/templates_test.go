package templates

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFuncMapHelpers(t *testing.T) {
	t.Parallel()
	fm := FuncMap()
	t.Run("default", func(t *testing.T) {
		t.Parallel()
		defaultFn := fm["default"].(func(any, any) any)
		require.Equal(t, "fallback", defaultFn("fallback", ""))
	})
	t.Run("coalesce", func(t *testing.T) {
		t.Parallel()
		coalesce := fm["coalesce"].(func(...any) any)
		require.Equal(t, "first", coalesce("first", "second"))
		require.Equal(t, "second", coalesce("", "second"))
	})
	t.Run("ensurePrefix", func(t *testing.T) {
		t.Parallel()
		ensure := fm["ensurePrefix"].(func(string, string) string)
		require.Equal(t, "#value", ensure("#", "value"))
		require.Equal(t, "", ensure("#", ""))
	})
	t.Run("ago", func(t *testing.T) {
		t.Parallel()
		ago := fm["ago"].(func(time.Time) string)
		require.Equal(t, "never", ago(time.Time{}))
	})
}

func TestParseStringTemplate(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		tmpl, err := ParseStringTemplate("name", "{{ .Value }}")
		require.NoError(t, err)

		res, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, "ok", res)
	})

	t.Run("error on empty input", func(t *testing.T) {
		_, err := ParseStringTemplate("empty", "")
		require.Error(t, err)
	})
}

func TestTemplateRender(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		tmpl, err := ParseStringTemplate("bytes", "{{ .Value }}")
		require.NoError(t, err)
		str, err := tmpl.Render(map[string]string{"Value": "payload"})
		require.NoError(t, err)
		require.Equal(t, "payload", str)

		tpl := Template{tmpl: tmpl.tmpl}
		output, err := tpl.Render(map[string]string{"Value": "payload"})
		require.NoError(t, err)
		require.Equal(t, []byte("payload"), output)
	})
	t.Run("error on empty input", func(t *testing.T) {
		t.Parallel()
		_, err := ParseStringTemplate("name", "")
		require.Error(t, err)
	})
}

func TestLoadString(t *testing.T) {
	t.Parallel()
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		_, err := LoadString("")
		require.Error(t, err)
	})
}

func TestLoadStringFromFS(t *testing.T) {
	t.Parallel()
	t.Run("error on empty input", func(t *testing.T) {
		t.Parallel()
		_, err := LoadStringFromFS(fstest.MapFS{}, "")
		require.Error(t, err)
	})
}

func TestLoadFromFS(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		fs := fstest.MapFS{
			"tmpl.tmpl": {Data: []byte("{{ .Value }}")},
		}
		tmpl, err := LoadFromFS(fs, "tmpl.tmpl")
		require.NoError(t, err)
		out, err := tmpl.Render(map[string]string{"Value": "ok"})
		require.NoError(t, err)
		require.Equal(t, []byte("ok"), out)
	})
}
