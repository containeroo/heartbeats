package handlers

import (
	"net/http"
	"strings"

	"github.com/containeroo/heartbeats/internal"
	log "github.com/sirupsen/logrus"
)

// Home is the handler for the / endpoint
func Home(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/heartbeats.html",
		"web/templates/footer.html",
	}
	ParseTemplates(templs, internal.HeartbeatsServer, w)
}
