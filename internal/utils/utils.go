package utils

import "reflect"

// DefaultIfZero returns fallback when value is the zero value.
func DefaultIfZero[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}
	return value
}

// ToPtr returns a pointer to the provided value.
func ToPtr[T any](value T) *T {
	return &value
}

// IsZeroValue returns true if the value is the zero value.
func IsZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	return rv.IsZero()
}
