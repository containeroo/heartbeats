package msteams

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Send_Integration(t *testing.T) {
	t.Parallel()

	var captured MSTeams
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // nolint:errcheck
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	client := New()
	_, err := client.Send(context.Background(), MSTeams{
		Title: "deploy",
		Text:  "done",
	}, srv.URL)

	assert.NoError(t, err)
	assert.Equal(t, "deploy", captured.Title)
	assert.Equal(t, "done", captured.Text)
}

func TestClient_Send_Integration_Non200(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	t.Cleanup(srv.Close)

	client := New()
	_, err := client.Send(context.Background(), MSTeams{}, srv.URL)

	assert.EqualError(t, err, "permanent: msteams http status: 400: bad")
}
