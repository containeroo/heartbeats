package notifier

import (
	"testing"
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

func TestFuncMap(t *testing.T) {
	t.Parallel()

	t.Run("upper", func(t *testing.T) {
		t.Parallel()

		fm := FuncMap()
		fn := fm["upper"].(func(string) string)
		assert.Equal(t, "HELLO", fn("hello"))
	})

	t.Run("lower", func(t *testing.T) {
		t.Parallel()

		fm := FuncMap()
		fn := fm["lower"].(func(string) string)
		assert.Equal(t, "hello", fn("HELLO"))
	})

	t.Run("formatTime", func(t *testing.T) {
		t.Parallel()

		fm := FuncMap()
		fn := fm["formatTime"].(func(time.Time, string) string)
		tm := time.Date(2023, 5, 1, 12, 34, 56, 0, time.UTC)
		assert.Equal(t, "2023-05-01", fn(tm, "2006-01-02"))
	})

	t.Run("ago non-zero", func(t *testing.T) {
		t.Parallel()

		fm := FuncMap()
		fn := fm["ago"].(func(time.Time) string)
		tm := time.Now().Add(-2 * time.Second)
		out := fn(tm)
		assert.NotEmpty(t, out)
		assert.NotEqual(t, "never", out)
	})

	t.Run("ago zero", func(t *testing.T) {
		t.Parallel()

		fm := FuncMap()
		fn := fm["ago"].(func(time.Time) string)
		assert.Equal(t, "never", fn(time.Time{}))
	})

	t.Run("join", func(t *testing.T) {
		t.Parallel()

		fm := FuncMap()
		fn := fm["join"].(func([]string, string) string)
		assert.Equal(t, "a-b-c", fn([]string{"a", "b", "c"}, "-"))
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
