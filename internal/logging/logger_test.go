package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	t.Run("JSON logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatJSON, false, &buf)

		logger.Info("test message", "key", "value")

		var logEntry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("expected JSON log, got error: %v", err)
		}

		if logEntry["msg"] != "test message" {
			t.Errorf("expected 'test message', got: %v", logEntry["msg"])
		}
		if logEntry["key"] != "value" {
			t.Errorf("expected 'key=value', got: %v", logEntry["key"])
		}
	})

	t.Run("Text logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatText, false, &buf)

		logger.Info("hello world", "foo", "bar")

		logOutput := buf.String()
		if !strings.Contains(logOutput, "hello world") {
			t.Errorf("expected message 'hello world' in text output, got: %s", logOutput)
		}
		if !strings.Contains(logOutput, "foo=bar") {
			t.Errorf("expected attribute 'foo=bar' in text output, got: %s", logOutput)
		}
	})

	t.Run("Invalid format - default to json", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		logger := SetupLogger("invalid", false, &buf)

		logger.Info("fallback check", "k", "v")

		var logEntry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("expected JSON fallback log, got error: %v", err)
		}
		if logEntry["msg"] != "fallback check" {
			t.Errorf("expected fallback log message, got: %v", logEntry["msg"])
		}
	})

	t.Run("Debug level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatText, true, &buf)

		logger.Debug("debug enabled", "foo", "bar")

		logOutput := buf.String()
		if !strings.Contains(logOutput, "debug enabled") {
			t.Errorf("expected debug message in output, got: %s", logOutput)
		}
	})
}
