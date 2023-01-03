package handlers

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal"
)

// Home is the handler for the / endpoint
func Home(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/heartbeats.html",
		"web/templates/footer.html",
	}
	ParseTemplates("base", templs, internal.HeartbeatsServer, w)
}
