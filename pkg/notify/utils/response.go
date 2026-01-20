package utils

import (
	"io"
	"net/url"
	"regexp"
	"strings"
)

// MaxResponseBody limits how much of a response body is read for errors/decoding.
const MaxResponseBody = 64 * 1024

// ReadBodyLimited reads at most MaxResponseBody bytes from r.
func ReadBodyLimited(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, MaxResponseBody))
}

var bearerRE = regexp.MustCompile(`(?i)Bearer\s+[^\s"']+`)

// RedactSecrets removes obvious secrets (tokens, query params) from a string.
func RedactSecrets(s string) string {
	s = bearerRE.ReplaceAllString(s, "Bearer ****")
	return redactURLs(s)
}

// redactURLs removes user info and query params from URLs.
func redactURLs(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s
	}
	for i, part := range parts {
		if !strings.HasPrefix(part, "http://") && !strings.HasPrefix(part, "https://") {
			continue
		}
		u, err := url.Parse(part)
		if err != nil {
			continue
		}
		u.User = nil
		u.RawQuery = ""
		u.Fragment = ""
		parts[i] = u.String()
	}
	return strings.Join(parts, " ")
}
