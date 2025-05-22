package msteams

import (
	"context"
	"encoding/json"
	"fmt"
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
	// Parameters:
	//   - ctx: Context for timeout or cancellation.
	//   - message: Structured payload with title and text.
	//   - webhookURL: Microsoft Teams webhook destination.
	//
	// Returns:
	//   - string: Success confirmation message.
	//   - error: If sending or response fails.
	Send(ctx context.Context, message MSTeams, webhookURL string) (string, error)
}

// Client sends messages to Microsoft Teams via webhook.
type Client struct {
	HttpClient utils.HTTPDoer // HttpClient executes the underlying HTTP request.
}

// New creates a new MS Teams client with optional headers and TLS control.
//
// Parameters:
//   - headers: Optional HTTP headers (e.g. Authorization, Content-Type).
//   - skipTLS: If true, disables TLS certificate validation.
//
// Returns:
//   - *Client: A ready-to-use Teams sender.
func New(headers map[string]string, skipTLS bool) *Client {
	return &Client{
		HttpClient: utils.NewHttpClient(headers, skipTLS),
	}
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
