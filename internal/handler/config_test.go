package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	t.Run("Returns config payload", func(t *testing.T) {
		t.Parallel()
		api := &API{
			Version: "v1",
			Commit:  "c1",
			SiteURL: "http://example.com",
		}

		h := api.Config()
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload configResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		assert.Equal(t, "v1", payload.Version)
		assert.Equal(t, "c1", payload.Commit)
		assert.Equal(t, "http://example.com", payload.SiteURL)
	})
}
