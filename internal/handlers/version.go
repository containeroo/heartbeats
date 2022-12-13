package handlers

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal"
)

// Version is the handler for the /version endpoint
func Version(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	outputFormat := GetOutputFormat(req)

	WriteOutput(w, http.StatusOK, outputFormat, &internal.HeartbeatsServer, "{{ .Version }}")
}
