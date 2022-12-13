package handlers

import (
	"net/http"
)

// Healthz is the handler for the /healthz endpoint
func Healthz(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	outputFormat := GetOutputFormat(req)

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}
