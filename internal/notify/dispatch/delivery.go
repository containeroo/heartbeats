package dispatch

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notify/targets"
	"github.com/containeroo/heartbeats/internal/notify/types"
)

// deliveryEngine is a notification delivery engine.
type deliveryEngine struct {
	logger  *slog.Logger
	history history.Recorder
	metrics *metrics.Registry
}

// NewDelivery constructs a Delivery with the provided logger.
func NewDelivery(
	logger *slog.Logger,
	historyStore history.Recorder,
	metricsReg *metrics.Registry,
) types.Delivery {
	return &deliveryEngine{
		logger:  logger,
		history: historyStore,
		metrics: metricsReg,
	}
}

// Dispatch sends a notification payload to all receivers.
func (d *deliveryEngine) Dispatch(
	ctx context.Context,
	payload types.Payload,
	receivers []*types.Receiver,
) error {
	var errs []error
	for _, rcv := range receivers {
		if err := d.dispatchReceiver(ctx, rcv, payload); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// dispatchReceiver sends a notification to a single receiver.
func (d *deliveryEngine) dispatchReceiver(
	ctx context.Context,
	rcv *types.Receiver,
	payload types.Payload,
) error {
	if rcv == nil {
		return errors.New("receiver is nil")
	}
	var errs []error
	for _, target := range rcv.Targets {
		var result types.DeliveryResult
		destination := targetDestination(target)
		attempts, err := withRetry(ctx, d.logger, rcv.Retry, func() error {
			if responder, ok := target.(types.ResultTarget); ok {
				res, sendErr := responder.SendResult(payload)
				result = res
				return sendErr
			}
			return target.Send(payload)
		})
		if err != nil {
			errs = append(errs, err)
			d.metrics.SetReceiverStatus(rcv.Name, target.Type(), payload.HeartbeatID, metrics.ERROR)
			d.logger.Debug("Notification target failed",
				"event", logging.EventNotificationTargetFailed.String(),
				"receiver", rcv.Name,
				"target_type", target.Type(),
				"error", err,
			)
			d.history.Add(history.Event{
				Type:        history.EventNotificationFailed.String(),
				HeartbeatID: payload.HeartbeatID,
				Receiver:    rcv.Name,
				TargetType:  target.Type(),
				Status:      payload.Status,
				Message:     err.Error(),
				Fields: map[string]any{
					"attempts":    attempts,
					"status":      result.Status,
					"status_code": result.StatusCode,
					"response":    result.Response,
					"target":      destination,
				},
			})
			continue
		}
		d.metrics.SetReceiverStatus(rcv.Name, target.Type(), payload.HeartbeatID, metrics.SUCCESS)
		d.logger.Debug("Notification target delivered",
			"event", logging.EventNotificationTargetDelivered.String(),
			"receiver", rcv.Name,
			"target_type", target.Type(),
		)
		d.history.Add(history.Event{
			Type:        history.EventNotificationDelivered.String(),
			HeartbeatID: payload.HeartbeatID,
			Receiver:    rcv.Name,
			TargetType:  target.Type(),
			Status:      payload.Status,
			Fields: map[string]any{
				"attempts":    attempts,
				"status":      result.Status,
				"status_code": result.StatusCode,
				"response":    result.Response,
				"target":      destination,
			},
		})
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func targetDestination(target types.Target) string {
	switch t := target.(type) {
	case *targets.WebhookTarget:
		return t.URL
	case *targets.EmailTarget:
		return strings.Join(t.To, ", ")
	default:
		return ""
	}
}
