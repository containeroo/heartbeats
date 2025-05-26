package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChain(t *testing.T) {
	t.Parallel()

	t.Run("no middleware", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		wrapped := Chain(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTeapot, rec.Code)
	})

	t.Run("single middleware", func(t *testing.T) {
		t.Parallel()

		var called bool

		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := Chain(handler, mw)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)
		assert.True(t, called)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("multiple middleware", func(t *testing.T) {
		t.Parallel()

		trace := ""

		mw1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace += "1"
				next.ServeHTTP(w, r)
			})
		}

		mw2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace += "2"
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace += "H"
			w.WriteHeader(http.StatusNoContent)
		})

		wrapped := Chain(handler, mw1, mw2)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, "21H", trace)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
