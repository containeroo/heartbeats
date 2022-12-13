package handlers

import (
	"net/http"
	"strings"

	"github.com/containeroo/heartbeats/internal"
	log "github.com/sirupsen/logrus"
)

// Version is the handler for the /version endpoint
func Version(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	WriteOutput(w, http.StatusOK, outputFormat, &internal.HeartbeatsServer, "{{ .Version }}")
}
