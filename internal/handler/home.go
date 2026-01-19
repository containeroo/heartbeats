package handler

import (
	"html/template"
	"net/http"
	"path"
)

// HomeHandler renders the base template with navbar and footer.
func (a *API) HomeHandler() http.HandlerFunc {
	tmpl := template.Must(
		template.New("base.html").
			ParseFS(a.webFS,
				path.Join("web/templates", "base.html"),
				path.Join("web/templates", "navbar.html"),
				path.Join("web/templates", "footer.html"),
			),
	)
	data := struct {
		Version string
		Commit  string
	}{
		Version: a.Version,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// execute the "base" template
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			a.respondJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		}
	}
}
