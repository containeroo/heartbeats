package slack

import (
	"bytes"
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

	var captured Slack
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // nolint:errcheck
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	client := New(WithEndpoint(srv.URL))
	_, err := client.Send(context.Background(), Slack{
		Channel: "#alerts",
		Attachments: []Attachment{{
			Color: "good",
			Title: "ok",
			Text:  "hello",
		}},
	})

	assert.NoError(t, err)
	assert.Equal(t, "#alerts", captured.Channel)
	if assert.Len(t, captured.Attachments, 1) {
		assert.Equal(t, "ok", captured.Attachments[0].Title)
		assert.Equal(t, "hello", captured.Attachments[0].Text)
	}
}

func TestClient_Send_Integration_Non200(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.Copy(w, bytes.NewBufferString("nope"))
	}))
	t.Cleanup(srv.Close)

	client := New(WithEndpoint(srv.URL))
	_, err := client.Send(context.Background(), Slack{})

	assert.EqualError(t, err, "permanent: slack http status: 418")
}
