package msteamsgraph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_SendChannel_Integration(t *testing.T) {
	t.Parallel()

	var captured Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // nolint:errcheck
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"123"}`))
	}))
	t.Cleanup(srv.Close)

	client := New(WithEndpointBase(srv.URL))
	_, err := client.SendChannel(context.Background(), "team", "channel", Message{
		Body: ItemBody{ContentType: "html", Content: "hi"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "html", captured.Body.ContentType)
	assert.Equal(t, "hi", captured.Body.Content)
}

func TestClient_SendChannel_Integration_Non200(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"BadRequest","message":"nope"}}`))
	}))
	t.Cleanup(srv.Close)

	client := New(WithEndpointBase(srv.URL))
	_, err := client.SendChannel(context.Background(), "team", "channel", Message{})

	assert.EqualError(t, err, "permanent: msteamsgraph http status: BadRequest: nope")
}
