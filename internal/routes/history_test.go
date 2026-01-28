package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/stretchr/testify/require"
)

func TestHistoryMiddleware_NilRecorder(t *testing.T) {
	t.Parallel()

	mw := HistoryMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/heartbeat/api", nil)
	recorder := httptest.NewRecorder()
	mw.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusAccepted, recorder.Result().StatusCode)
}

func TestHistoryMiddleware_RecordsEvent(t *testing.T) {
	t.Parallel()

	store := history.NewStore(5)
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	mw := HistoryMiddleware(store, inner)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/api", nil)
	req.SetPathValue("id", "api")
	recorder := httptest.NewRecorder()
	start := time.Now()
	mw.ServeHTTP(recorder, req)

	require.True(t, called)
	res := recorder.Result()
	require.Equal(t, http.StatusTeapot, res.StatusCode)

	events := store.List()
	require.Len(t, events, 1)
	ev := events[0]
	require.Equal(t, history.EventHTTPAccess.String(), ev.Type)
	require.Equal(t, "api", ev.HeartbeatID)
	require.Equal(t, "http_access", ev.Message)
	require.Equal(t, req.Method, ev.Fields["method"])
	require.Equal(t, req.URL.Path, ev.Fields["url_path"])
	require.Equal(t, http.StatusTeapot, ev.Fields["status_code"])
	require.GreaterOrEqual(t, ev.Time.UnixNano(), start.UnixNano())
	require.NotEmpty(t, ev.Fields["duration"])
}
