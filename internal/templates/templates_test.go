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
		require.Equal(t, "value", defaultFn("fallback", "value"))
	})

	t.Run("coalesce", func(t *testing.T) {
		t.Parallel()

		coalesce := fm["coalesce"].(func(...any) any)

		require.Equal(t, "first", coalesce("first", "second"))
		require.Equal(t, "second", coalesce("", "second"))
		require.Nil(t, coalesce("", 0, false, nil))
	})

	t.Run("ensurePrefix", func(t *testing.T) {
		t.Parallel()

		ensure := fm["ensurePrefix"].(func(string, any) string)

		require.Equal(t, "#value", ensure("#", "value"))
		require.Equal(t, "#value", ensure("#", "#value"))
		require.Equal(t, "", ensure("#", ""))
		require.Equal(t, "", ensure("#", nil))
	})

	t.Run("toUpper", func(t *testing.T) {
		t.Parallel()

		toUpper := fm["toUpper"].(func(any) string)

		require.Equal(t, "VALUE", toUpper("value"))
	})

	t.Run("toLower", func(t *testing.T) {
		t.Parallel()

		toLower := fm["toLower"].(func(any) string)

		require.Equal(t, "value", toLower("VALUE"))
	})

	t.Run("formatTime", func(t *testing.T) {
		t.Parallel()

		formatTime := fm["formatTime"].(func(any, string) (string, error))
		value := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)

		got, err := formatTime(value, "2006-01-02 15:04:05 MST")

		require.NoError(t, err)
		require.Equal(t, "2026-06-24 12:00:00 UTC", got)
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
		t.Parallel()

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

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		_, err := LoadString("")

		require.Error(t, err)
	})
}

func TestLoadStringFromFS(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		fs := fstest.MapFS{
			"string.tmpl": {Data: []byte("{{ .Value }}")},
		}

		tmpl, err := LoadStringFromFS(fs, "string.tmpl")
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

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		_, err := LoadFromFS(fstest.MapFS{}, "")

		require.Error(t, err)
	})
}
