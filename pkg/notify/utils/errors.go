package utils

import (
	"errors"
	"fmt"
)

// ErrorKind identifies whether an error is transient or permanent.
type ErrorKind string

const (
	ErrorTransient ErrorKind = "transient"
	ErrorPermanent ErrorKind = "permanent"
)

// Error wraps an error with a kind and operation.
type Error struct {
	Kind ErrorKind
	Op   string
	Err  error
}

func (e *Error) Error() string {
	if e.Op == "" {
		return fmt.Sprintf("%s: %v", e.Kind, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Kind, e.Op, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// Wrap annotates err with kind and op.
func Wrap(kind ErrorKind, op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Kind: kind, Op: op, Err: err}
}

// IsTransient returns true if err is a transient error.
func IsTransient(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind == ErrorTransient
	}
	return false
}

// KindFromStatus maps HTTP status to error kind.
func KindFromStatus(status int) ErrorKind {
	switch {
	case status == 408 || status == 429 || status >= 500:
		return ErrorTransient
	default:
		return ErrorPermanent
	}
}
