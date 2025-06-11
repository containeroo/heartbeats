package utils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoRequest(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		// Set up a mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/test", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"hello":"world"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`ok`)) // nolint:errcheck
		}))
		defer server.Close()

		headers := map[string]string{
			"Content-Type": "application/json",
		}
		client := NewHttpClient(headers, false)

		resp, err := client.DoRequest(context.Background(), "POST", server.URL+"/test", []byte(`{"hello":"world"}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close() // nolint:errcheck
		assert.NoError(t, err)
		assert.Equal(t, "ok", string(data))
	})

	t.Run("Invalid method", func(t *testing.T) {
		t.Parallel()

		client := NewHttpClient(nil, false)

		// Trigger error: invalid method (contains space)
		resp, err := client.DoRequest(context.Background(), "INVALID METHOD", "http://example.com", nil)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error creating INVALID METHOD request")
	})
}
