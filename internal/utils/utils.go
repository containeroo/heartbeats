package utils

// DefaultIfZero returns fallback when value is the zero value.
func DefaultIfZero[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}
	return value
}
