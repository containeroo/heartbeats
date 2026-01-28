package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithCategory(t *testing.T) {
	t.Parallel()

	t.Run("nil logger uses default", func(t *testing.T) {
		t.Parallel()

		logger := WithCategory(nil, CategorySystem)
		assert.NotNil(t, logger)
	})

	t.Run("empty category returns logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		out := WithCategory(logger, "")

		out.Info("hello")
		assert.NotContains(t, buf.String(), "category=")
	})

	t.Run("category added", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		out := WithCategory(logger, CategoryAccess)

		out.Info("hello")
		assert.Contains(t, buf.String(), "category=access")
	})
}

func TestCategoryLoggers(t *testing.T) {
	t.Parallel()

	t.Run("access logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		AccessLogger(logger).Info("access")
		assert.Contains(t, buf.String(), "category=access")
	})

	t.Run("business logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		BusinessLogger(logger).Info("biz")
		assert.Contains(t, buf.String(), "category=business")
	})

	t.Run("system logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		SystemLogger(logger).Info("system")
		assert.Contains(t, buf.String(), "category=system")
	})
}
