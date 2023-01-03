package handlers

import (
	"net/http"
)

// NotFound is the handler for the 404 page
func NotFound(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	if outputFormat := req.URL.Query().Get("output"); outputFormat != "" {
		// if output is given, return history in wanted format
		WriteOutput(w, http.StatusNotFound, outputFormat, &ResponseStatus{Status: "nok", Error: "404 Not Found"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/404.html",
		"web/templates/footer.html",
	}
	ParseTemplates("base", templs, nil, w)
}
