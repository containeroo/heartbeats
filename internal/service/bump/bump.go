package bump

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
)

// ErrUnknownHeartbeat signals that the heartbeat ID is not registered.
var ErrUnknownHeartbeat = errors.New("unknown heartbeat id")

// Receive records a heartbeat bump and updates the in-memory state.
func Receive(
	ctx context.Context,
	mgr *heartbeat.Manager,
	hist history.Store,
	logger *slog.Logger,
	id string,
	source string,
	method string,
	userAgent string,
) error {
	if mgr.Get(id) == nil {
		logger.Warn("received bump for unknown heartbeat ID", "id", id, "from", source)
		return fmt.Errorf("%w: %s", ErrUnknownHeartbeat, id)
	}

	logger.Info("received bump", "id", id, "from", source)

	payload := history.RequestMetadataPayload{
		Source:    source,
		Method:    method,
		UserAgent: userAgent,
	}
	ev := history.MustNewEvent(history.EventTypeHeartbeatReceived, id, payload)

	if err := hist.Append(ctx, ev); err != nil {
		logger.Error("failed to record state change", "err", err)
		return err
	}

	// We check if the heartbeat exists before calling Receive.
	mgr.Receive(id) // nolint:errcheck

	return nil
}

// Fail records a manual failure and updates the in-memory state.
func Fail(
	ctx context.Context,
	mgr *heartbeat.Manager,
	hist history.Store,
	logger *slog.Logger,
	id string,
	source string,
	method string,
	userAgent string,
) error {
	if mgr.Get(id) == nil {
		logger.Warn("received /fail bump for unknown heartbeat ID", "id", id, "from", source)
		return fmt.Errorf("%w: %s", ErrUnknownHeartbeat, id)
	}

	logger.Info("manual fail", "id", id, "from", source)

	payload := history.RequestMetadataPayload{
		Source:    source,
		Method:    method,
		UserAgent: userAgent,
	}
	ev := history.MustNewEvent(history.EventTypeHeartbeatFailed, id, payload)

	if err := hist.Append(ctx, ev); err != nil {
		logger.Error("failed to record state change", "err", err)
		return err
	}

	// We check if the heartbeat exists before calling Fail.
	mgr.Fail(id) // nolint:errcheck

	return nil
}
