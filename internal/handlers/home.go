package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
	"path"
)

// HomeHandler renders the base template with navbar and footer.
func HomeHandler(staticFS fs.FS, version, commit string) http.HandlerFunc {
	tmpl := template.Must(
		template.New("base.html").
			ParseFS(staticFS,
				path.Join("web/templates", "base.html"),
				path.Join("web/templates", "navbar.html"),
				path.Join("web/templates", "footer.html"),
			),
	)
	data := struct {
		Version string
		Commit  string
	}{
		Version: version,
		Commit:  commit,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// execute the "base" template
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
