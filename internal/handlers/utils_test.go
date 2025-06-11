package handlers

import (
	"testing"

	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

func TestMasqueradeURL(t *testing.T) {
	t.Parallel()

	t.Run("https URL with long path", func(t *testing.T) {
		t.Parallel()

		raw := "https://outlook.office.com/webhook/abc123456789xyz"
		masked := masqueradeURL(raw, 6)

		assert.Equal(t, "https://outlook.office.com/...789xyz", masked)
	})

	t.Run("http URL with short path", func(t *testing.T) {
		t.Parallel()

		raw := "http://example.com/x"
		masked := masqueradeURL(raw, 4)

		assert.Equal(t, "http://example.com/...om/x", masked)
	})

	t.Run("URL shorter than tailLen", func(t *testing.T) {
		t.Parallel()

		raw := "https://short.io/x"
		masked := masqueradeURL(raw, 100)

		assert.Equal(t, "https://short.io/...https://short.io/x", masked)
	})

	t.Run("malformed URL returns fallback", func(t *testing.T) {
		t.Parallel()

		masked := masqueradeURL("::::", 4)
		assert.Equal(t, "<invalid-url>", masked)
	})
}

func TestFormatEmailRecipients(t *testing.T) {
	t.Parallel()

	t.Run("only To", func(t *testing.T) {
		t.Parallel()

		out := formatEmailRecipients(notifier.EmailDetails{To: []string{"a@example.com"}})
		assert.Equal(t, "To: a@example.com", out)
	})

	t.Run("To, CC, BCC", func(t *testing.T) {
		t.Parallel()

		out := formatEmailRecipients(notifier.EmailDetails{
			To:  []string{"a@example.com"},
			CC:  []string{"c@example.com"},
			BCC: []string{"b@example.com"},
		})
		assert.Equal(t, "To: a@example.com\nCC: c@example.com\nBCC: b@example.com", out)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		out := formatEmailRecipients(notifier.EmailDetails{})
		assert.Equal(t, "", out)
	})
}
