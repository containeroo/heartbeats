package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestHomeHandler(t *testing.T) {
	t.Parallel()

	t.Run("renders base template with version", func(t *testing.T) {
		t.Parallel()

		// in-memory file system with minimal template files
		webFS := fstest.MapFS{
			"web/templates/base.html":   &fstest.MapFile{Data: []byte(`{{define "base"}}<html>{{template "navbar"}}<footer>{{.Version}}</footer>{{end}}`)},
			"web/templates/navbar.html": &fstest.MapFile{Data: []byte(`{{define "navbar"}}<nav>nav</nav>{{end}}`)},
			"web/templates/footer.html": &fstest.MapFile{Data: []byte(`{{define "footer"}}<!-- footer -->{{end}}`)},
		}

		version := "v1.2.3"
		handler := HomeHandler(webFS, version)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close() // nolint:errcheck

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := rec.Body.String()
		assert.Contains(t, body, version)
		assert.Contains(t, body, "<nav>nav</nav>")
		assert.Contains(t, body, "<footer>"+version+"</footer>")
	})

	t.Run("parse error", func(t *testing.T) {
		t.Parallel()

		// in-memory file system with minimal template files
		webFS := fstest.MapFS{
			"web/templates/base.html":   &fstest.MapFile{Data: []byte(``)},
			"web/templates/navbar.html": &fstest.MapFile{Data: []byte(`{{define "navbar"}}<nav>nav</nav>{{end}}`)},
			"web/templates/footer.html": &fstest.MapFile{Data: []byte(`{{define "footer"}}<!-- footer -->{{end}}`)},
		}

		version := "v1.2.3"
		handler := HomeHandler(webFS, version)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		resp := rec.Result()
		defer resp.Body.Close() // nolint:errcheck

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, "html/template: \"base\" is undefined\n", rec.Body.String())
	})
}
