package logging

import "log/slog"

const (
	// CategoryAccess tags access logs.
	CategoryAccess = "access"
	// CategoryBusiness tags domain logs.
	CategoryBusiness = "business"
	// CategorySystem tags system logs.
	CategorySystem = "system"
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
func AccessLogger(logger *slog.Logger) *slog.Logger {
	return WithCategory(logger, CategoryAccess)
}

// BusinessLogger returns a logger tagged for business logs.
func BusinessLogger(logger *slog.Logger) *slog.Logger {
	return WithCategory(logger, CategoryBusiness)
}

// SystemLogger returns a logger tagged for system logs.
func SystemLogger(logger *slog.Logger) *slog.Logger {
	return WithCategory(logger, CategorySystem)
}
