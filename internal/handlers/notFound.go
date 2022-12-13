package handlers

import "net/http"

// NotFound is the handler for the 404 page
func NotFound(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/404.html",
		"web/templates/footer.html",
	}
	ParseTemplates("base", templs, nil, w)
}
