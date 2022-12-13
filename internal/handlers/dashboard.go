package handlers

import (
	"net/http"
	"strings"

	"github.com/containeroo/heartbeats/internal"
	log "github.com/sirupsen/logrus"
)

// HandlerDashboard is the handler for the /dashboard endpoint
func Dashboard(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	templs := []string{
		"web/templates/dashboard.html",
	}

	ParseTemplates("dashboard", templs, &internal.HeartbeatsServer, w)
}
