package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/containeroo/heartbeats/pkg/notify/utils"
)

const slackAPIEndpoint = "https://slack.com/api/chat.postMessage" // Slack API endpoint for posting messages

// Slack defines the payload structure for a Slack message.
type Slack struct {
	Channel     string       `json:"channel"`     // Channel is the Slack channel where the message will be posted.
	Attachments []Attachment `json:"attachments"` // Attachments is the list of structured message blocks.
}

// Attachment represents a visual block attached to a Slack message.
type Attachment struct {
	Color string `json:"color"` // Color is the left border color of the attachment.
	Text  string `json:"text"`  // Text is the body content of the attachment.
	Title string `json:"title"` // Title is the heading text of the attachment.
}

// Response represents the structure of a Slack API response.
type Response struct {
	Ok    bool   `json:"ok"`    // Ok is true if the request was successful.
	Error string `json:"error"` // Error is a human-readable error message, if any.
}

// Sender defines the interface for sending Slack messages.
type Sender interface {
	// Send posts a Slack message to the configured API endpoint.
	//
	// Parameters:
	//   - ctx: Context to control timeout or cancellation of the request.
	//   - slackMessage: The structured Slack message payload to send.
	//
	// Returns:
	//   - *Response: The Slack API response object.
	//   - error: Any error that occurred during sending or decoding.
	Send(ctx context.Context, slackMessage Slack) (*Response, error)
}

// Client sends Slack messages using the official chat.postMessage API.
type Client struct {
	HttpClient utils.HTTPDoer // HttpClient is used to send HTTP requests (mockable for testing).
}

// New creates a new Slack API client with optional headers and TLS settings.
//
// Parameters:
//   - headers: Custom HTTP headers, including the Authorization bearer token.
//   - skipTLS: Whether to skip TLS certificate validation.
//
// Returns:
//   - *Client: A configured Slack client implementing the Sender interface.
func New(headers map[string]string, skipTLS bool) *Client {
	return &Client{
		HttpClient: utils.NewHttpClient(headers, skipTLS),
	}
}

// NewWithToken returns a Slack API client using the provided bearer token.
//
// Parameters:
//   - token: Slack API token to use for authentication.
//   - skipTLS: If true, disables TLS certificate verification.
//
// Returns:
//   - *Client: A configured Slack client ready to send messages.
func NewWithToken(token string, skipTLS bool) *Client {
	return New(map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}, skipTLS)
}

// Send posts a message to the Slack chat.postMessage endpoint.
//
// Parameters:
//   - ctx: Context to manage request lifetime.
//   - slackMessage: Message payload including channel and attachments.
//
// Returns:
//   - *Response: Structured API response containing status and errors.
//   - error: If sending or decoding fails, or Slack returns a failure.
func (c *Client) Send(ctx context.Context, slackMessage Slack) (*Response, error) {
	// Encode the Slack message to JSON
	data, err := json.Marshal(slackMessage)
	if err != nil {
		return nil, fmt.Errorf("error marshalling Slack message: %w", err)
	}

	// Send the HTTP request using the injected HTTP client
	resp, err := c.HttpClient.DoRequest(ctx, "POST", slackAPIEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	// Slack returns 200 OK even for some failures—check explicitly
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var parsed Response
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if !parsed.Ok {
		return &parsed, fmt.Errorf("Slack API error: %s", parsed.Error)
	}

	return &parsed, nil
}
