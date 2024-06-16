package server

import (
	"context"
	"embed"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	version := "1.0.0"

	log := logger.NewLogger(true)

	heartbeatStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()
	historyStore := history.NewStore()

	config := Config{
		ListenAddress: "localhost:8080",
		SiteRoot:      "http://localhost:8080",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	var templates embed.FS
	err := Run(
		ctx,
		log,
		version,
		config,
		templates,
		heartbeatStore,
		notificationStore,
		historyStore,
	)
	assert.NoError(t, err)
}
