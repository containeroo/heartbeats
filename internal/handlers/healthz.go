package handlers

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// HeartbeatsServer is the handler for the /healthz endpoint
func Healthz(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}
