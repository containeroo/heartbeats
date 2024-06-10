package handlers

import (
	"heartbeats/internal/logger"
	"net/http"
)

func NotFound(logger logger.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			body := `
<html>
	<head>
		<title>404 Not Found</title>
	</head>
	<body>
		<h1>404 Not Found</h1>
		<p>The page you requested could not be found.</p>
	</body>
</html>
	`
			if _, err := w.Write([]byte(body)); err != nil {
				logger.Errorf("Error writing 404 response: %v", err)
			}
		})
}
