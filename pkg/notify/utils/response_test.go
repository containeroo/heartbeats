package utils

import (
	"strings"
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	t.Parallel()

	input := "Bearer abc123 https://example.com/path?token=secret#frag ok"
	got := RedactSecrets(input)

	if got == input {
		t.Fatalf("expected redaction, got unchanged: %q", got)
	}
	if strings.Contains(got, "abc123") {
		t.Fatalf("expected token to be redacted, got: %q", got)
	}
	if strings.Contains(got, "token=secret") {
		t.Fatalf("expected query to be stripped, got: %q", got)
	}
}
