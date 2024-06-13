package server

import (
	"context"
	"embed"
	"heartbeats/pkg/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	log := logger.NewLogger(true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	var staticFS embed.FS
	err := Run(ctx, "localhost:8080", staticFS, log)
	assert.NoError(t, err)
}
