package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/assert"
)

func setupRouter(t *testing.T, hbName string, hist history.Store) http.Handler {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	heartbeats := heartbeat.HeartbeatConfigMap{
		hbName: {
			ID:          hbName,
			Description: hbName,
			Interval:    30 * time.Second,
			Grace:       10 * time.Second,
			Receivers:   []string{"rec"},
		},
	}

	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	recorder := servicehistory.NewRecorder(hist)
	metricsReg := metrics.New(hist)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, metricsReg)
	factory := heartbeat.DefaultActorFactory{
		Logger:     logger,
		History:    recorder,
		Metrics:    metricsReg,
		DispatchCh: disp.Mailbox(),
	}
	mgr, err := heartbeat.NewManagerFromHeartbeatMap(
		context.Background(),
		heartbeats,
		logger,
		factory,
	)
	assert.NoError(t, err)
	api := NewAPI(
		"test",
		"test",
		nil,
		"",
		"",
		true,
		logger,
		mgr,
		hist,
		recorder,
		disp,
		nil,
		nil,
	)
	router := http.NewServeMux()
	router.Handle("GET /no-id/", api.BumpHandler())
	router.Handle("GET /no-id/fail", api.FailHandler())
	router.Handle("GET /bump/{id}", api.BumpHandler())
	router.Handle("POST /bump/{id}", api.BumpHandler())
	router.Handle("GET /bump/{id}/fail", api.FailHandler())
	router.Handle("POST /bump/{id}/fail", api.FailHandler())
	return router
}

func TestBumpHandler(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/no-id/", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "missing id", resp.Error)
	})

	t.Run("not found id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bump/not-found", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "unknown heartbeat id \"not-found\"", resp.Error)
	})

	t.Run("successful bump GET", func(t *testing.T) {
		t.Parallel()

		hbName := "get-success"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/bump/%s", hbName), nil)

		req.RemoteAddr = "1.2.3.4:5678"
		req.Header.Set("User-Agent", "Go-test")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp statusResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "ok", resp.Status)

		e := hist.ListByID(hbName)

		ev := e[0]
		assert.Equal(t, history.EventTypeHeartbeatReceived, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)

		var meta history.RequestMetadataPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &meta))
		assert.Equal(t, "GET", meta.Method)
		assert.Equal(t, "1.2.3.4:5678", meta.Source)
		assert.Equal(t, "Go-test", meta.UserAgent)
	})

	t.Run("successful bump POST", func(t *testing.T) {
		t.Parallel()

		hbName := "post-success"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/bump/%s", hbName), nil)
		req.RemoteAddr = "5.6.7.8:1234"
		req.Header.Set("User-Agent", "Go-post")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp statusResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "ok", resp.Status)

		e := hist.ListByID(hbName)
		ev := e[0]

		assert.Equal(t, history.EventTypeHeartbeatReceived, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)

		var meta history.RequestMetadataPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &meta))
		assert.Equal(t, "POST", meta.Method)
	})
}

func TestFailHandler(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/no-id/fail", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "missing id", resp.Error)
	})

	t.Run("not found id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bump/not-found/fail", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "unknown heartbeat id \"not-found\"", resp.Error)
	})

	t.Run("successful fail GET", func(t *testing.T) {
		t.Parallel()

		hbName := "check-success-fail"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/bump/%s/fail", hbName), nil)
		req.RemoteAddr = "1.2.3.4:5678"
		req.Header.Set("User-Agent", "Go-test")

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp statusResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "ok", resp.Status)

		events := hist.ListByID(hbName)
		assert.Len(t, events, 1)

		ev := events[0]
		assert.Equal(t, history.EventTypeHeartbeatFailed, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)

		var meta history.RequestMetadataPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &meta))

		assert.Equal(t, "GET", meta.Method)
		assert.Equal(t, "1.2.3.4:5678", meta.Source)
		assert.Equal(t, "Go-test", meta.UserAgent)
	})

	t.Run("successful fail POST", func(t *testing.T) {
		t.Parallel()

		hbName := "post-success-fail"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/bump/%s/fail", hbName), nil)
		req.RemoteAddr = "5.6.7.8:1234"
		req.Header.Set("User-Agent", "Go-post")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp statusResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "ok", resp.Status)

		e := hist.ListByID(hbName)
		ev := e[0]

		assert.Equal(t, history.EventTypeHeartbeatFailed, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)

		var meta history.RequestMetadataPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &meta))
		assert.Equal(t, "POST", meta.Method)
	})

	t.Run("bump - record event error", func(t *testing.T) {
		t.Parallel()

		mockHist := &history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		hbName := "event-error"
		router := setupRouter(t, hbName, mockHist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/bump/%s", hbName), nil)
		req.RemoteAddr = "5.6.7.8:1234"
		req.Header.Set("User-Agent", "Go-post")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "fail!", resp.Error)
	})

	t.Run("fail - record event error", func(t *testing.T) {
		t.Parallel()

		mockHist := &history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		hbName := "event-error"
		router := setupRouter(t, hbName, mockHist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/bump/%s/fail", hbName), nil)
		req.RemoteAddr = "5.6.7.8:1234"
		req.Header.Set("User-Agent", "Go-post")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "fail!", resp.Error)
	})
}
