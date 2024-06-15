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
	log := logger.NewLogger(true)

	heartbeatStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()
	historyStore := history.NewStore()
	version := "1.0.0"
	listenAdderss := "localhost:8080"
	siteRoot := "http://localhost:8080"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	var templates embed.FS
	err := Run(ctx, listenAdderss, version, siteRoot, templates, log, heartbeatStore, notificationStore, historyStore)
	assert.NoError(t, err)
}
