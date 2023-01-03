package handlers

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal"
)

// HandlerDashboard is the handler for the /dashboard endpoint
func Dashboard(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	templs := []string{
		"web/templates/dashboard.html",
	}

	ParseTemplates("dashboard", templs, &internal.HeartbeatsServer, w)
}
