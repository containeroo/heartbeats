package handlers

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// HandlerDashboard is the handler for the /dashboard endpoint
func Dashboard(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	templs := []string{
		"web/templates/dashboard.html",
	}

	ParseTemplates(templs, w)
}
