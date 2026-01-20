package msteams

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient is a test double for utils.HTTPDoer
type mockHTTPClient struct {
	Response *http.Response
	Err      error
	Capture  []byte
}

func (m *mockHTTPClient) DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	m.Capture = body
	return m.Response, m.Err
}

func TestClient_Send_Success(t *testing.T) {
	t.Parallel()

	payload := MSTeams{
		Title: "System Alert",
		Text:  "Something happened",
	}

	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		},
	}

	client := &Client{HttpClient: mock}
	result, err := client.Send(context.Background(), payload, "https://example.com/webhook")

	assert.NoError(t, err)
	assert.Equal(t, "Message sent successfully", result)
}

func TestClient_Send_RequestFailure(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Err: io.ErrUnexpectedEOF,
	}

	client := &Client{HttpClient: mock}
	_, err := client.Send(context.Background(), MSTeams{}, "https://webhook")

	assert.Error(t, err)
	assert.EqualError(t, err, "transient: msteams request: unexpected EOF")
}

func TestClient_Send_Non200Status(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(bytes.NewBufferString("forbidden")),
		},
	}

	client := &Client{HttpClient: mock}
	_, err := client.Send(context.Background(), MSTeams{}, "https://webhook")

	assert.Error(t, err)
	assert.EqualError(t, err, "permanent: msteams http status: 403: forbidden")
}
