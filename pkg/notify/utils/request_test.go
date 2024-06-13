package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpClient(t *testing.T) {
	t.Run("Create new HttpClient", func(t *testing.T) {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		client := NewHttpClient(headers, true)

		assert.NotNil(t, client)
		assert.Equal(t, headers, client.Headers)
		assert.True(t, client.SkipInsecure)
	})
}

func TestHttpClient_DoRequest(t *testing.T) {
	t.Run("Perform successful GET request", func(t *testing.T) {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		client := NewHttpClient(headers, true)

		// Setup a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		ctx := context.Background()
		resp, err := client.DoRequest(ctx, http.MethodGet, server.URL, nil)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Perform GET request with invalid URL", func(t *testing.T) {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		client := NewHttpClient(headers, true)

		ctx := context.Background()
		_, err := client.DoRequest(ctx, http.MethodGet, "http://invalid-url", nil)

		assert.Error(t, err)
	})
}

func TestHttpClient_createHTTPClient(t *testing.T) {
	t.Run("Create HTTP client with custom transport", func(t *testing.T) {
		client := NewHttpClient(nil, true)
		httpClient := client.createHTTPClient()

		assert.NotNil(t, httpClient)
		assert.IsType(t, &http.Client{}, httpClient)
		assert.IsType(t, &http.Transport{}, httpClient.Transport)

		transport := httpClient.Transport.(*http.Transport)
		assert.NotNil(t, transport.TLSClientConfig)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}
