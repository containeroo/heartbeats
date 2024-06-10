package handlers

import "time"

// isFalse returns true if the given boolean pointer is nil or false.
func isFalse(b *bool) bool {
	return b != nil && !*b
}

// isTrue returns true if the given boolean pointer is not nil and true.
func isTrue(b *bool) bool {
	return b != nil && *b
}

// formatTime formats the given time with the given format
func formatTime(t time.Time, format string) string {
	// check if the time is zero or time is not set
	if t.IsZero() || t.Unix() == 0 {
		return "-"
	}
	return t.Format(format)
}
