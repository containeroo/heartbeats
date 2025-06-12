package msteams

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/containeroo/heartbeats/pkg/notify/utils"
)

// MSTeams defines the payload structure for a Microsoft Teams message.
type MSTeams struct {
	Title string `json:"title"` // Title is the main heading of the message card.
	Text  string `json:"text"`  // Text is the body content of the message card.
}

// Sender is implemented by types capable of sending Microsoft Teams messages.
type Sender interface {
	// Send transmits the message to the given MS Teams webhook.
	//
	// Returns a confirmation or error on failure.
	Send(ctx context.Context, message MSTeams, webhookURL string) (string, error)
}

// Client sends messages to Microsoft Teams via webhook.
type Client struct {
	HttpClient utils.HTTPDoer // HttpClient executes the underlying HTTP request.
}

// Option configures a MSTeams client.
type Option func(*Client)

// WithHeaders sets additional HTTP headers for the MS Teams client.
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		hc := c.HttpClient.(*utils.HttpClient)
		if hc.Headers == nil {
			hc.Headers = make(map[string]string)
		}
		maps.Copy(hc.Headers, headers)
	}
}

// WithInsecureTLS sets whether to skip TLS certificate verification.
func WithInsecureTLS(skipInsecure bool) Option {
	return func(c *Client) {
		if hc, ok := c.HttpClient.(*utils.HttpClient); ok {
			hc.SkipInsecure = skipInsecure
		}
	}
}

// New creates a new MS Teams client with functional options.
//
// Use options like WithHeaders or WithInsecureTLS to customize behavior.
func New(opts ...Option) *Client {
	client := &Client{
		HttpClient: utils.NewHttpClient(nil, false),
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// Send delivers the provided message to the specified Teams webhook URL.
//
// Parameters:
//   - ctx: Request-scoped context.
//   - message: Message payload with title and text.
//   - webhookURL: Fully qualified MS Teams webhook endpoint.
//
// Returns:
//   - string: Success text if response is HTTP 200.
//   - error: If marshalling, sending, or decoding fails.
func (c *Client) Send(ctx context.Context, message MSTeams, webhookURL string) (string, error) {
	// Serialize the message to JSON.
	data, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("error marshalling MS Teams message: %w", err)
	}

	// Execute the POST request using the configured HTTP client.
	resp, err := c.HttpClient.DoRequest(ctx, "POST", webhookURL, data)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	// Verify the HTTP status code is 200 OK.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	return "Message sent successfully", nil
}
