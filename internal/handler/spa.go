package handler

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"path"
	"text/template"
	"time"
)

// SPA serves the built React bundle from frontend/dist.
func SPA(spaFS fs.FS, routePrefix string) http.HandlerFunc {
	indexFile, err := spaFS.Open("index.html")
	if err != nil {
		// Defer runtime error until handler execution so startup still works without a build.
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "app not built (run npm run build in frontend)", http.StatusServiceUnavailable)
		}
	}
	defer indexFile.Close() // nolint:errcheck

	indexData, err := io.ReadAll(indexFile)
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unable to read spa bundle", http.StatusInternalServerError)
		}
	}

	indexInfo, _ := indexFile.Stat()
	indexMod := infoModTime(indexInfo)
	rawIndex := string(indexData)

	tmpl, err := template.New("index").Parse(rawIndex)
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unable to parse spa template", http.StatusInternalServerError)
		}
	}

	baseHref := "/"
	if routePrefix != "" {
		baseHref = routePrefix + "/"
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]string{
		"RoutePrefix": routePrefix,
		"BaseHref":    baseHref,
	}); err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unable to render spa template", http.StatusInternalServerError)
		}
	}
	indexBytes := rendered.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, path.Base("index.html"), indexMod, bytes.NewReader(indexBytes))
	}
}

// infoModTime guards against nil file info.
func infoModTime(info fs.FileInfo) time.Time {
	if info == nil {
		return time.Time{}
	}
	return info.ModTime()
}
