package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsFalse(t *testing.T) {
	bTrue := true
	bFalse := false

	assert.True(t, IsFalse(&bFalse), "Expected isFalse to return true for false pointer")
	assert.False(t, IsFalse(&bTrue), "Expected isFalse to return false for true pointer")
	assert.False(t, IsFalse(nil), "Expected isFalse to return false for nil pointer")
}

func TestIsTrue(t *testing.T) {
	bTrue := true
	bFalse := false

	assert.False(t, IsTrue(&bFalse), "Expected isTrue to return false for false pointer")
	assert.True(t, IsTrue(&bTrue), "Expected isTrue to return true for true pointer")
	assert.False(t, IsTrue(nil), "Expected isTrue to return false for nil pointer")
}

func TestFormatTime(t *testing.T) {
	format := "2006-01-02 15:04:05"

	t.Run("Non-zero time", func(t *testing.T) {
		tm := time.Date(2021, 9, 15, 14, 0, 0, 0, time.UTC)
		expected := "2021-09-15 14:00:00"
		assert.Equal(t, expected, FormatTime(tm, format), "Expected formatted time to match")
	})

	t.Run("Zero time", func(t *testing.T) {
		tm := time.Time{}
		expected := "-"
		assert.Equal(t, expected, FormatTime(tm, format), "Expected formatted time to be '-' for zero time")
	})
}

func TestIsRecent(t *testing.T) {
	t.Run("Time is within a second", func(t *testing.T) {
		now := time.Now()
		assert.True(t, IsRecent(now), "Expected isRecent to return true for current time")
	})

	t.Run("Time is not within a second", func(t *testing.T) {
		past := time.Now().Add(-2 * time.Second)
		assert.False(t, IsRecent(past), "Expected isRecent to return false for time more than a second ago")
	})

	t.Run("Time is exactly one second ago", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Second)
		assert.False(t, IsRecent(past), "Expected isRecent to return false for time exactly one second ago")
	})

	t.Run("Time is 999 milliseconds ago", func(t *testing.T) {
		past := time.Now().Add(-999 * time.Millisecond)
		assert.True(t, IsRecent(past), "Expected isRecent to return true for time 999 milliseconds ago")
	})

	t.Run("Time is 1001 milliseconds ago", func(t *testing.T) {
		past := time.Now().Add(-1001 * time.Millisecond)
		assert.False(t, IsRecent(past), "Expected isRecent to return false for time 1001 milliseconds ago")
	})
}
