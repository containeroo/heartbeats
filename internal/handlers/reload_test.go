package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReloadHandler(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

	api := NewAPI(
		"test",
		"test",
		nil,
		"",
		"",
		false,
		logger,
		nil,
		nil,
		nil,
		nil,
		nil,
		func() error {
			calls.Add(1)
			return nil
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/-/reload", nil)
	rec := httptest.NewRecorder()
	api.ReloadHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp statusResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, int32(1), calls.Load())
}
