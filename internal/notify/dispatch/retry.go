package dispatch

import (
	"context"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notify/types"
)

// withRetry executes fn with retry attempts and delay.
func withRetry(ctx context.Context, logger *slog.Logger, cfg types.RetryConfig, fn func() error) (int, error) {
	attempts := max(cfg.Count+1, 1)
	var lastErr error
	attempt := 0
	for attempt := range attempts {
		if attempt > 0 && cfg.Delay > 0 {
			logger.Debug("Notification target dispatch",
				"event", logging.EventNotificationTargetDispatch.String(),
				"attempt", attempt+1,
			)
			select {
			case <-ctx.Done():
				return attempt + 1, ctx.Err() // context cancelled
			case <-time.After(cfg.Delay): // wait before next retry
			}
		}
		if err := fn(); err != nil {
			lastErr = err
			continue
		}
		return attempt + 1, nil
	}
	return attempt + 1, lastErr
}
