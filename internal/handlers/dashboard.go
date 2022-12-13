package handlers

import (
	"net/http"
	"strings"

	"github.com/containeroo/heartbeats/internal/docs"
	log "github.com/sirupsen/logrus"
)

// Dashboard is the handler for the dashboard page
func Dashboard(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	templs := []string{
		"web/templates/dashboard.html",
	}

	ParseTemplates(templs, &docs.Documentation, w)
}
