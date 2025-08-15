package routeutil

import (
	"net/url"
	"strings"
)

// NormalizeRoutePrefix returns "" or "/prefix" from input, accepting raw paths or full URLs.
func NormalizeRoutePrefix(input string) string {
	s := strings.TrimSpace(input)
	if s == "" || s == "/" {
		return ""
	}
	// If someone passes a full URL, keep only the .Path.
	if strings.Contains(s, "://") {
		if u, err := url.Parse(s); err == nil {
			s = u.Path
		}
	}
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "/")
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if s == "/" {
		return ""
	}
	return s
}
