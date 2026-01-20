package logging

import (
	"context"
	"log/slog"
)

const (
	CategoryAccess   = "access"
	CategoryBusiness = "business"
	CategoryDB       = "db"
	CategorySystem   = "system"
)

// WithCategory adds a category field to the logger.
func WithCategory(logger *slog.Logger, category string) *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	if category == "" {
		return logger
	}
	return logger.With("category", category)
}

// AccessLogger returns a logger tagged for access logs.
func AccessLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	return WithCategory(WithRequestIDLogger(logger, ctx), CategoryAccess)
}

// BusinessLogger returns a logger tagged for business logs.
func BusinessLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	return WithCategory(WithRequestIDLogger(logger, ctx), CategoryBusiness)
}

// DBLogger returns a logger tagged for db logs.
func DBLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	return WithCategory(WithRequestIDLogger(logger, ctx), CategoryDB)
}

// SystemLogger returns a logger tagged for system logs.
func SystemLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	return WithCategory(WithRequestIDLogger(logger, ctx), CategorySystem)
}
