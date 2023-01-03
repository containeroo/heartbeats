package handlers

import (
	"net/http"
)

// Healthz is the handler for the /healthz endpoint
func Healthz(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	WriteOutput(w, http.StatusOK, GetOutputFormat(req), &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}
