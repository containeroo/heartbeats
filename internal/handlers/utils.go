package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/containeroo/heartbeats/internal/notifier"
)

// masqueradeURL hides most of a URL, showing only scheme, host, and last N characters of path.
func masqueradeURL(raw string, tailLen int) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid-url>"
	}

	full := u.String()
	tail := full
	if len(full) > tailLen {
		tail = full[len(full)-tailLen:]
	}
	return fmt.Sprintf("%s://%s/...%s", u.Scheme, u.Host, tail)
}

// formatEmailRecipients returns a string representation of all email recipients.
func formatEmailRecipients(details notifier.EmailDetails) string {
	parts := []string{}

	if len(details.To) > 0 {
		parts = append(parts, "To: "+strings.Join(details.To, ", "))
	}
	if len(details.CC) > 0 {
		parts = append(parts, "CC: "+strings.Join(details.CC, ", "))
	}
	if len(details.BCC) > 0 {
		parts = append(parts, "BCC: "+strings.Join(details.BCC, ", "))
	}

	return strings.Join(parts, "\n")
}
