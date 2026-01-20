package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"time"

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
	Endpoint   string         // Endpoint is the Slack API endpoint to use.
}

// Option configures a Slack client.
type Option func(*Client)

// WithHeaders sets additional HTTP headers for the Slack client.
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
		c.HttpClient.(*utils.HttpClient).SkipInsecure = skipInsecure
	}
}

// WithEndpoint overrides the Slack API endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.Endpoint = endpoint
	}
}

// WithTimeout sets the per-request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.HttpClient.(*utils.HttpClient).Timeout = timeout
	}
}

// New creates a Slack API client with optional configuration.
//
// Options can be used to set headers or disable TLS verification.
func New(opts ...Option) *Client {
	// start with empty headers + default TLS settings
	client := &Client{
		HttpClient: utils.NewHttpClient(make(map[string]string), false),
		Endpoint:   slackAPIEndpoint,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// NewWithToken returns a Slack client configured with the given bearer token.
//
// Options can be passed to customize behavior (e.g. headers, TLS).
//   - token: Slack API token for authentication.
//   - opts: Optional functional options (e.g. WithInsecureTLS, WithHeaders).
func NewWithToken(token string, opts ...Option) *Client {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	return New(append([]Option{WithHeaders(headers)}, opts...)...)
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
		return nil, utils.Wrap(utils.ErrorPermanent, "slack marshal", err)
	}

	// Send the HTTP request using the injected HTTP client
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = slackAPIEndpoint
	}
	resp, err := c.HttpClient.DoRequest(ctx, "POST", endpoint, data)
	if err != nil {
		return nil, utils.Wrap(utils.ErrorTransient, "slack request", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	// Slack returns 200 OK even for some failures—check explicitly
	if resp.StatusCode != http.StatusOK {
		return nil, utils.Wrap(utils.KindFromStatus(resp.StatusCode), "slack http status", fmt.Errorf("%d", resp.StatusCode))
	}

	var parsed Response
	if err := json.NewDecoder(io.LimitReader(resp.Body, utils.MaxResponseBody)).Decode(&parsed); err != nil {
		return nil, utils.Wrap(utils.ErrorPermanent, "slack decode", err)
	}

	if !parsed.Ok {
		return &parsed, utils.Wrap(utils.ErrorPermanent, "slack api", fmt.Errorf("%s", parsed.Error))
	}

	return &parsed, nil
}
