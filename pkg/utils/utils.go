package utils

import (
	"time"
)

// IsFalse returns true if the given boolean pointer is nil or false.
func IsFalse(b *bool) bool {
	return b != nil && !*b
}

// IsTrue returns true if the given boolean pointer is not nil and true.
func IsTrue(b *bool) bool {
	return b != nil && *b
}

// FormatTime formats the given time with the given format
func FormatTime(t time.Time, format string) string {
	// check if the time is zero or time is not set
	if t.IsZero() || t.Unix() == 0 {
		return "-"
	}

	return t.Format(format)
}

// IsRecent checks if the given time is within the last second
func IsRecent(t time.Time) bool {
	return time.Since(t) < time.Second
}
