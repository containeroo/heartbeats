package notifier

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplytemplate(t *testing.T) {
	t.Parallel()

	t.Run("valid template", func(t *testing.T) {
		t.Parallel()

		data := struct{ ID string }{ID: "abc"}
		out, err := applyTemplate("hello {{ .ID }}", data)
		assert.NoError(t, err)
		assert.Equal(t, "hello abc", out)
	})

	t.Run("invalid template parse", func(t *testing.T) {
		t.Parallel()

		_, err := applyTemplate("{{ .ID ", struct{ ID string }{ID: "abc"})
		assert.Error(t, err)
	})

	t.Run("invalid execution", func(t *testing.T) {
		t.Parallel()

		_, err := applyTemplate("{{ .Missing }}", struct{ ID string }{ID: "abc"})
		assert.Error(t, err)
	})
}

func render(t *testing.T, tpl string, data any) string {
	t.Helper()
	tmpl := template.Must(template.New("test").Funcs(FuncMap()).Parse(tpl))
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	assert.NoError(t, err)
	return buf.String()
}

func TestFuncMapTemplateUsage(t *testing.T) {
	t.Parallel()

	t.Run("upper", func(t *testing.T) {
		t.Parallel()
		out := render(t, `{{ "foo" | upper }}`, nil)
		assert.Equal(t, "FOO", out)
	})

	t.Run("lower", func(t *testing.T) {
		t.Parallel()
		out := render(t, `{{ "BaR" | lower }}`, nil)
		assert.Equal(t, "bar", out)
	})

	t.Run("formatTime", func(t *testing.T) {
		t.Parallel()
		now := time.Date(2025, 6, 24, 12, 0, 0, 0, time.UTC)
		out := render(t, `{{ formatTime .Time "2006-01-02" }}`, map[string]any{"Time": now})
		assert.Equal(t, "2025-06-24", out)
	})

	t.Run("isRecent true", func(t *testing.T) {
		t.Parallel()
		out := render(t, `{{ isRecent .Time }}`, map[string]any{"Time": time.Now()})
		assert.Equal(t, "true", out)
	})

	t.Run("isRecent false", func(t *testing.T) {
		t.Parallel()
		out := render(t, `{{ isRecent .Time }}`, map[string]any{"Time": time.Now().Add(-10 * time.Second)})
		assert.Equal(t, "false", out)
	})

	t.Run("ago non-zero", func(t *testing.T) {
		t.Parallel()
		start := time.Now().Add(-2 * time.Second)
		out := render(t, `{{ ago .Time }}`, map[string]any{"Time": start})
		assert.Contains(t, out, "2s")
	})

	t.Run("ago zero", func(t *testing.T) {
		t.Parallel()
		out := render(t, `{{ ago .Time }}`, map[string]any{"Time": time.Time{}})
		assert.Equal(t, "never", out)
	})

	t.Run("join", func(t *testing.T) {
		t.Parallel()
		out := render(t, `{{ join .Items ", " }}`, map[string]any{"Items": []string{"a", "b", "c"}})
		assert.Equal(t, "a, b, c", out)
	})
}

func TestResolveSkipTLS(t *testing.T) {
	t.Parallel()

	t.Run("explicit true", func(t *testing.T) {
		t.Parallel()

		val := true
		assert.True(t, resolveSkipTLS(&val, false))
	})

	t.Run("explicit false", func(t *testing.T) {
		t.Parallel()

		val := false
		assert.False(t, resolveSkipTLS(&val, true))
	})

	t.Run("explicit nil", func(t *testing.T) {
		t.Parallel()

		assert.True(t, resolveSkipTLS(nil, true))
		assert.False(t, resolveSkipTLS(nil, false))
	})
}

func TestFormatNotification(t *testing.T) {
	t.Parallel()

	t.Run("uses defaults", func(t *testing.T) {
		t.Parallel()

		data := NotificationData{ID: "abc"}
		out, err := formatNotification(data, "", "", "Title: {{ .ID }}", "Body")

		assert.NoError(t, err)
		assert.Equal(t, "Title: abc", out.Title)
		assert.Equal(t, "Body", out.Message)
	})

	t.Run("uses custom templates", func(t *testing.T) {
		t.Parallel()

		data := NotificationData{ID: "ping-42", Status: "grace", LastBump: time.Now().Add(-2 * time.Second)}
		out, err := formatNotification(data, "[{{ .Status }}] {{ .ID }}", "last: {{ ago .LastBump }}", "fallback", "fallback")

		assert.NoError(t, err)
		assert.Equal(t, "[grace] ping-42", out.Title)
		assert.NotEmpty(t, out.Message)
	})

	t.Run("fails on bad template", func(t *testing.T) {
		t.Parallel()

		data := NotificationData{ID: "x"}
		_, err := formatNotification(data, "{{ .Missing }}", "ok", "fallback", "fallback")

		assert.Error(t, err)
	})
}

func TestFormatEmailRecipients(t *testing.T) {
	t.Parallel()

	t.Run("only To", func(t *testing.T) {
		t.Parallel()

		out := FormatEmailRecipients(EmailDetails{To: []string{"a@example.com"}})
		assert.Equal(t, "To: a@example.com", out)
	})

	t.Run("To, CC, BCC", func(t *testing.T) {
		t.Parallel()

		out := FormatEmailRecipients(EmailDetails{
			To:  []string{"a@example.com"},
			CC:  []string{"c@example.com"},
			BCC: []string{"b@example.com"},
		})
		assert.Equal(t, "To: a@example.com\nCC: c@example.com\nBCC: b@example.com", out)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		out := FormatEmailRecipients(EmailDetails{})
		assert.Equal(t, "", out)
	})
}

func TestMasqueradeURL(t *testing.T) {
	t.Parallel()

	t.Run("https URL with long path", func(t *testing.T) {
		t.Parallel()

		raw := "https://outlook.office.com/webhook/abc123456789xyz"
		masked := MasqueradeURL(raw, 6)

		assert.Equal(t, "https://outlook.office.com/...789xyz", masked)
	})

	t.Run("http URL with short path", func(t *testing.T) {
		t.Parallel()

		raw := "http://example.com/x"
		masked := MasqueradeURL(raw, 4)

		assert.Equal(t, "http://example.com/...om/x", masked)
	})

	t.Run("URL shorter than tailLen", func(t *testing.T) {
		t.Parallel()

		raw := "https://short.io/x"
		masked := MasqueradeURL(raw, 100)

		assert.Equal(t, "https://short.io/...https://short.io/x", masked)
	})

	t.Run("malformed URL returns fallback", func(t *testing.T) {
		t.Parallel()

		masked := MasqueradeURL("::::", 4)
		assert.Equal(t, "<invalid-url>", masked)
	})
}
