package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsFalse(t *testing.T) {
	bTrue := true
	bFalse := false

	assert.True(t, isFalse(&bFalse), "Expected isFalse to return true for false pointer")
	assert.False(t, isFalse(&bTrue), "Expected isFalse to return false for true pointer")
	assert.False(t, isFalse(nil), "Expected isFalse to return false for nil pointer")
}

func TestIsTrue(t *testing.T) {
	bTrue := true
	bFalse := false

	assert.False(t, isTrue(&bFalse), "Expected isTrue to return false for false pointer")
	assert.True(t, isTrue(&bTrue), "Expected isTrue to return true for true pointer")
	assert.False(t, isTrue(nil), "Expected isTrue to return false for nil pointer")
}

func TestFormatTime(t *testing.T) {
	format := "2006-01-02 15:04:05"

	t.Run("Non-zero time", func(t *testing.T) {
		tm := time.Date(2021, 9, 15, 14, 0, 0, 0, time.UTC)
		expected := "2021-09-15 14:00:00"
		assert.Equal(t, expected, formatTime(tm, format), "Expected formatted time to match")
	})

	t.Run("Zero time", func(t *testing.T) {
		tm := time.Time{}
		expected := "-"
		assert.Equal(t, expected, formatTime(tm, format), "Expected formatted time to be '-' for zero time")
	})
}

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
