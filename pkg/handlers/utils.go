package handlers

import (
	"net/http"
	"strings"
)

// getClientIP extracts the client's IP address from the request, preferring X-Forwarded-For if present.
func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	return r.RemoteAddr
}
