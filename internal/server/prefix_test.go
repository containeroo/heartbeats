package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMountUnderPrefix(t *testing.T) {
	t.Parallel()

	// Inner mux that our wrapper will mount under a prefix.
	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "root") // nolint:errcheck
		// Prove we reached the inner handler at "/".
	})
	inner.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "foo") // nolint:errcheck
		// Prove path "/foo" is recognized after StripPrefix.
	})
	inner.HandleFunc("GET /api/v1/ok", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok") // nolint:errcheck
		// Prove deeper paths still work.
	})

	t.Run("empty prefix returns original handler (serves at root)", func(t *testing.T) {
		t.Parallel()

		h := mountUnderPrefix(inner, "") // no prefix → unchanged

		// Request to root should be handled by inner mux' "/" handler.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "root", rec.Body.String())

		// With empty prefix, "/tiledash/foo" is ALSO matched by "/" (catch-all) → 200 "root".
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/tiledash/foo", nil)
		h.ServeHTTP(rec2, req2)
		require.Equal(t, http.StatusOK, rec2.Code)
		require.Equal(t, "root", rec2.Body.String())
	})

	t.Run("bare prefix redirects to prefix with trailing slash", func(t *testing.T) {
		t.Parallel()

		h := mountUnderPrefix(inner, "/tiledash")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tiledash", nil) // bare prefix without slash
		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/tiledash/", rec.Header().Get("Location"))
	})

	t.Run("prefixed paths are stripped and routed to inner handler", func(t *testing.T) {
		t.Parallel()

		h := mountUnderPrefix(inner, "/tiledash")

		// Request under the prefix should be stripped and served by inner "/foo".
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tiledash/foo", nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "foo", rec.Body.String())

		// Deeper nested path under the prefix should also be routed correctly.
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/tiledash/api/v1/ok", nil)
		h.ServeHTTP(rec2, req2)
		require.Equal(t, http.StatusOK, rec2.Code)
		assert.Equal(t, "ok", rec2.Body.String())
	})

	t.Run("non-prefixed paths 404 when mounted under a prefix", func(t *testing.T) {
		t.Parallel()

		h := mountUnderPrefix(inner, "/tiledash")

		// Direct root path should not be served when app is mounted under /tiledash only.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}
