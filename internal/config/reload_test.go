package config

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatchReload(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ch := make(chan os.Signal, 1)
	done := make(chan struct{})

	go WatchReload(ctx, ch, logger, func() error {
		close(done)
		return nil
	})

	ch <- os.Interrupt

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		require.Fail(t, "timeout waiting for reload")
	}
}
