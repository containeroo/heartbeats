package logger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogrusLogger(t *testing.T) {
	var buf bytes.Buffer

	t.Run("NewLogger sets correct log level", func(t *testing.T) {
		log := NewLogger(true)
		assert.Equal(t, logrus.DebugLevel, log.logger.Level, "Expected log level to be Debug")

		log = NewLogger(false)
		assert.Equal(t, logrus.InfoLevel, log.logger.Level, "Expected log level to be Info")
	})

	t.Run("SetOutput redirects log output", func(t *testing.T) {
		log := NewLogger(false)
		log.SetOutput(&buf)

		log.Info("Test log")
		assert.Contains(t, buf.String(), "Test log", "Expected buffer to contain log message")
	})

	t.Run("Debug logs at Debug level", func(t *testing.T) {
		buf.Reset()
		log := NewLogger(true)
		log.SetOutput(&buf)

		log.Debug("Debug message")
		assert.Contains(t, buf.String(), "DEBUG", "Expected buffer to contain 'DEBUG'")
		assert.Contains(t, buf.String(), "Debug message", "Expected buffer to contain 'Debug message'")
	})

	t.Run("Info logs at Info level", func(t *testing.T) {
		buf.Reset()
		log := NewLogger(false)
		log.SetOutput(&buf)

		log.Info("Info message")
		assert.Contains(t, buf.String(), "INFO", "Expected buffer to contain 'INFO'")
		assert.Contains(t, buf.String(), "Info message", "Expected buffer to contain 'Info message'")
	})

	t.Run("Warn logs at Warn level", func(t *testing.T) {
		buf.Reset()
		log := NewLogger(false)
		log.SetOutput(&buf)

		log.Warn("Warn message")
		assert.Contains(t, buf.String(), "WARN", "Expected buffer to contain 'WARN'")
		assert.Contains(t, buf.String(), "Warn message", "Expected buffer to contain 'Warn message'")
	})

	t.Run("Error logs at Error level", func(t *testing.T) {
		buf.Reset()
		log := NewLogger(false)
		log.SetOutput(&buf)

		log.Error("Error message")
		assert.Contains(t, buf.String(), "ERROR", "Expected buffer to contain 'ERROR'")
		assert.Contains(t, buf.String(), "Error message", "Expected buffer to contain 'Error message'")
	})

	t.Run("Write logs message at specified level", func(t *testing.T) {
		buf.Reset()
		log := NewLogger(true)
		log.SetOutput(&buf)

		log.Write(DebugLevel, "Write Debug message")
		assert.Contains(t, buf.String(), "Write Debug message", "Expected buffer to contain 'Write Debug message'")

		buf.Reset()
		log.Write(InfoLevel, "Write Info message")
		assert.Contains(t, buf.String(), "Write Info message", "Expected buffer to contain 'Write Info message'")

		buf.Reset()
		log.Write(WarnLevel, "Write Warn message")
		assert.Contains(t, buf.String(), "Write Warn message", "Expected buffer to contain 'Write Warn message'")

		buf.Reset()
		log.Write(ErrorLevel, "Write Error message")
		assert.Contains(t, buf.String(), "Write Error message", "Expected buffer to contain 'Write Error message'")
	})

	t.Run("Writef logs formatted message at specified level", func(t *testing.T) {
		buf.Reset()
		log := NewLogger(true)
		log.SetOutput(&buf)

		log.Writef(DebugLevel, "Writef %s", "Debug message")
		assert.Contains(t, buf.String(), "Writef Debug message", "Expected buffer to contain 'Writef Debug message'")

		buf.Reset()
		log.Writef(InfoLevel, "Writef %s", "Info message")
		assert.Contains(t, buf.String(), "Writef Info message", "Expected buffer to contain 'Writef Info message'")

		buf.Reset()
		log.Writef(WarnLevel, "Writef %s", "Warn message")
		assert.Contains(t, buf.String(), "Writef Warn message", "Expected buffer to contain 'Writef Warn message'")

		buf.Reset()
		log.Writef(ErrorLevel, "Writef %s", "Error message")
		assert.Contains(t, buf.String(), "Writef Error message", "Expected buffer to contain 'Writef Error message'")
	})
}
