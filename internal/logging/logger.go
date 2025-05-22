package logging

import (
	"io"
	"log/slog"
)

// LogFormat defines the supported log formats.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// SetupLogger configures a structured logger with the specified format and debug mode.
func SetupLogger(format LogFormat, debug bool, output io.Writer) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}

	if debug {
		handlerOpts.Level = slog.LevelDebug
	}

	var handler slog.Handler
	switch format {
	case LogFormatJSON:
		handler = slog.NewJSONHandler(output, handlerOpts)
	case LogFormatText:
		handler = slog.NewTextHandler(output, handlerOpts)
	default:
		// Default to JSON if an invalid format is provided.
		handler = slog.NewJSONHandler(output, handlerOpts)
	}

	return slog.New(handler)
}
