package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultIfZero(t *testing.T) {
	t.Parallel()
	t.Run("String uses fallback on empty", func(t *testing.T) {
		t.Parallel()
		result := DefaultIfZero("", "fallback")
		assert.Equal(t, "fallback", result)
	})

	t.Run("String keeps value", func(t *testing.T) {
		t.Parallel()
		result := DefaultIfZero("value", "fallback")
		assert.Equal(t, "value", result)
	})

	t.Run("Int uses fallback on zero", func(t *testing.T) {
		t.Parallel()
		result := DefaultIfZero(0, 42)
		assert.Equal(t, 42, result)
	})

	t.Run("Int keeps value", func(t *testing.T) {
		t.Parallel()
		result := DefaultIfZero(7, 42)
		assert.Equal(t, 7, result)
	})

	t.Run("Time uses fallback on zero", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		result := DefaultIfZero(time.Time{}, now)
		assert.Equal(t, now, result)
	})

	t.Run("Time keeps value", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		result := DefaultIfZero(now, time.Now().Add(time.Minute))
		assert.Equal(t, now, result)
	})
}
