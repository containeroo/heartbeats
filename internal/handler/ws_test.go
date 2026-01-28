package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeHub struct {
	called bool
}

func (f *fakeHub) Handle(w http.ResponseWriter, r *http.Request) {
	f.called = true
	w.WriteHeader(http.StatusAccepted)
}

func TestWSHandler_NoHub(t *testing.T) {
	api := &API{}
	handler := api.WS()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	require.Contains(t, w.Body.String(), "websocket hub disabled")
}

func TestWSHandler_WithHub(t *testing.T) {
	hub := &fakeHub{}
	api := &API{wsHub: hub}
	handler := api.WS()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)

	handler.ServeHTTP(w, req)
	require.True(t, hub.called)
	require.Equal(t, http.StatusAccepted, w.Code)
}
