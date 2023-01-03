package handlers

import (
	"fmt"
	"net/http"

	"github.com/containeroo/heartbeats/internal/docs"
	"github.com/containeroo/heartbeats/internal/utils"
	"github.com/gorilla/mux"
)

// Docs is the handler for the docs page
func Docs(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/docs.html",
		"web/templates/docs/api.html",
		"web/templates/footer.html",
	}

	ParseTemplates("base", templs, &docs.Documentation, w)
}

// Chapter is the handler for the documentation chapters
func Chapter(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	vars := mux.Vars(req)
	chapter := vars["chapter"]

	var templs []string

	if !utils.IsInListOfStrings(docs.Chapters, chapter) {
		templs = []string{
			"web/templates/base.html",
			"web/templates/navbar.html",
			"web/templates/docs.html",
			"web/templates/docs/404.html",
			"web/templates/footer.html",
		}
	} else {
		templs = []string{
			"web/templates/base.html",
			"web/templates/navbar.html",
			"web/templates/docs.html",
			fmt.Sprintf("web/templates/docs/%s.html", chapter),
			"web/templates/footer.html",
		}
	}
	ParseTemplates("base", templs, &docs.Documentation, w)
}
