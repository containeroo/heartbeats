package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseRecorder_WriteHeader(t *testing.T) {
	t.Parallel()

	t.Run("captures written status code", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		rr := &responseRecorder{ResponseWriter: rec}
		assert.Equal(t, 0, rr.statusCode) // default uninitialized

		rr.WriteHeader(http.StatusNotFound)
		assert.Equal(t, http.StatusNotFound, rr.statusCode)
		assert.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	})
}
