package handlers

import "net/http"

func NotFound(w http.ResponseWriter, req *http.Request) {
	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/404.html",
		"web/templates/footer.html",
	}
	ParseTemplates(templs, w)
}
