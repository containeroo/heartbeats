package test

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/notifier"
)

// SendTestNotification sends a test notification to a specific receiver.
func SendTestNotification(dispatcher *notifier.Dispatcher, logger *slog.Logger, id string) {
	logger.Info("Test request received", "receiver", id)

	dispatcher.Mailbox() <- notifier.NotificationData{
		ID:        fmt.Sprintf("manual-test-%s", time.Now().Format(time.RFC3339)),
		Receivers: []string{id},
		Title:     "Test Notification",
		Message:   "This is a test notification",
	}
}

// TriggerTestHeartbeat sends a test notification for a specific heartbeat.
func TriggerTestHeartbeat(mgr *heartbeat.Manager, logger *slog.Logger, id string) error {
	logger.Info("Test request heartbeat", "heartbeat", id)
	return mgr.Test(id)
}
