package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondJSON(t *testing.T) {
	t.Run("Writes JSON response", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		require.NoError(t, encode(rec, http.StatusCreated, map[string]string{"status": "ok"}))

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var payload map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "ok", payload["status"])
	})
}
