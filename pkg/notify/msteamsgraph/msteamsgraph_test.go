package msteamsgraph

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient simulates utils.HTTPDoer for testing.
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
	client := New(WithHeaders(map[string]string{"Authorization": "Bearer token"}), WithInsecureTLS(true))
	assert.IsType(t, &Client{}, client)
}

func TestNewWithToken(t *testing.T) {
	t.Parallel()
	client := NewWithToken("token", WithInsecureTLS(true))
	assert.IsType(t, &Client{}, client)
}

func TestClient_SendChannel_Success(t *testing.T) {
	t.Parallel()

	message := Message{
		Body: ItemBody{
			ContentType: "html",
			Content:     "<b>Test</b>",
		},
	}

	mockRespBody := `{"id":"abc123","webUrl":"https://teams.microsoft.com/l/message/abc123"}`
	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
		},
	}

	client := &Client{HttpClient: mock}
	resp, err := client.SendChannel(context.Background(), "team-id", "channel-id", message)

	assert.NoError(t, err)
	assert.Equal(t, "abc123", resp.ID)
	assert.Equal(t, string(mock.Capture), "{\"body\":{\"contentType\":\"html\",\"content\":\"\\u003cb\\u003eTest\\u003c/b\\u003e\"}}")
}

func TestClient_SendChannel_GraphError(t *testing.T) {
	t.Parallel()

	mockRespBody := `{"error":{"code":"InvalidAuthToken","message":"Access denied."}}`
	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
		},
	}

	client := &Client{HttpClient: mock}
	_, err := client.SendChannel(context.Background(), "team", "channel", Message{})

	assert.Error(t, err)
	assert.EqualError(t, err, "teams graph error: InvalidAuthToken: Access denied.")
}

func TestClient_SendChannel_DecodeError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("{invalid-json")),
		},
	}

	client := &Client{HttpClient: mock}
	_, err := client.SendChannel(context.Background(), "team", "channel", Message{})

	assert.Error(t, err)
	assert.EqualError(t, err, "decode error: invalid character 'i' looking for beginning of object key string")
}

func TestClient_SendChannel_RequestError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		Err: assert.AnError,
	}

	client := &Client{HttpClient: mock}
	_, err := client.SendChannel(context.Background(), "team", "channel", Message{})

	assert.Error(t, err)
	assert.EqualError(t, err, "request failed: assert.AnError general error for testing")
}
