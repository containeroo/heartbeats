package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	t.Run("With X-Forwarded-For header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
		expected := "192.168.1.1"
		assert.Equal(t, expected, getClientIP(req), "Expected to return the first IP from X-Forwarded-For header")
	})

	t.Run("Without X-Forwarded-For header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		expected := "127.0.0.1:8080"
		assert.Equal(t, expected, getClientIP(req), "Expected to return RemoteAddr when X-Forwarded-For header is absent")
	})
}
