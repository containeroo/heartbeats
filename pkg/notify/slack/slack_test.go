package slack

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient simulates utils.HTTPDoer for testing
type mockHTTPClient struct {
	Response *http.Response
	Err      error
	Capture  []byte
}

func (m *mockHTTPClient) DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	m.Capture = body
	return m.Response, m.Err
}

func TestNew(t *testing.T) {
	t.Parallel()

	client := New(WithHeaders(map[string]string{"Authorization": "Bearer my-token"}), WithInsecureTLS(true))
	assert.IsType(t, &Client{}, client)
}

func TestNewWithToken(t *testing.T) {
	t.Parallel()

	client := NewWithToken("my-token", WithInsecureTLS(true))
	assert.IsType(t, &Client{}, client)
}

func TestClient_Send_Success(t *testing.T) {
	t.Parallel()

	slackPayload := Slack{
		Channel: "#alerts",
		Attachments: []Attachment{{
			Color: "good",
			Title: "Test Alert",
			Text:  "Everything is operational.",
		}},
	}

	mockRespBody := `{"ok":true}`
	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
		},
	}

	client := &Client{HttpClient: mock}
	resp, err := client.Send(context.Background(), slackPayload)

	assert.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Contains(t, string(mock.Capture), "Everything is operational.")
}

func TestClient_Send_Non200(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(bytes.NewBufferString("Forbidden")),
		},
	}

	client := &Client{HttpClient: mock}
	_, err := client.Send(context.Background(), Slack{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "received non-200 response")
}

func TestClient_Send_ErrorDecoding(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("{invalid-json")),
		},
	}

	client := &Client{HttpClient: mock}
	_, err := client.Send(context.Background(), Slack{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding response")
}

func TestClient_Send_ErrorAPIResponse(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"ok":false,"error":"invalid_auth"}`)),
		},
	}

	client := &Client{HttpClient: mock}
	_, err := client.Send(context.Background(), Slack{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Slack API error: invalid_auth")
}
