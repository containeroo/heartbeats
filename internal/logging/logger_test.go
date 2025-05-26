package logging

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	t.Run("JSON logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatJSON, false, &buf)

		logger.Info("test message", "key", "value")

		var logEntry map[string]any
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "test message", logEntry["msg"])
		assert.Equal(t, "value", logEntry["key"])
	})

	t.Run("Text logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatText, false, &buf)

		logger.Info("hello world", "foo", "bar")

		logOutput := buf.String()

		assert.Contains(t, logOutput, "hello world")
		assert.Contains(t, logOutput, "foo=bar")
	})

	t.Run("Invalid format - default to json", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger("invalid", false, &buf)

		logger.Info("fallback check", "k", "v")

		var logEntry map[string]any
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "fallback check", logEntry["msg"])
	})

	t.Run("Debug level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatText, true, &buf)

		logger.Debug("debug enabled", "foo", "bar")

		logOutput := buf.String()
		assert.Contains(t, logOutput, `level=DEBUG msg="debug enabled" foo=bar`)
	})
}
