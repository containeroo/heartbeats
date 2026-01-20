package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
)

const RequestIDHeader = "X-Request-Id"

type requestIDKey struct{}

var requestIDReader io.Reader = rand.Reader

// NewRequestID generates a random request id.
func NewRequestID() string {
	var buf [16]byte
	if _, err := io.ReadFull(requestIDReader, buf[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buf[:])
}

// WithRequestID stores a request id in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID extracts a request id from the context.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// WithRequestIDLogger adds request_id to the logger if present in ctx.
func WithRequestIDLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	if id := RequestID(ctx); id != "" {
		return logger.With("request_id", id)
	}
	return logger
}
