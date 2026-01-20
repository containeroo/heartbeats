package msteamsgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/utils"
)

const defaultEndpointBase string = "https://graph.microsoft.com/v1.0"

// Message represents the payload structure for a Teams message.
type Message struct {
	Body ItemBody `json:"body"` // Body contains the message content.
}

// ItemBody wraps the message content in HTML format.
type ItemBody struct {
	ContentType string `json:"contentType"` // e.g., "html"
	Content     string `json:"content"`     // Message body
}

// Response represents a simplified Graph API response.
type Response struct {
	ID      string `json:"id,omitempty"`      // ID of the message
	WebURL  string `json:"webUrl,omitempty"`  // Web link to the posted message
	Error   *Error `json:"error,omitempty"`   // Optional error object
	Message string `json:"message,omitempty"` // Sometimes returned on failure
}

// Error represents Graph API error details.
type Error struct {
	Code       string `json:"code"`       // Error code
	Message    string `json:"message"`    // Error message
	InnerError any    `json:"innerError"` // Optional nested error
}

// Sender defines methods to send messages via Microsoft Graph.
type Sender interface {
	SendChannel(ctx context.Context, teamID, channelID string, msg Message) (*Response, error)
}

// Client posts messages using the Microsoft Graph API.
type Client struct {
	HttpClient utils.HTTPDoer // HTTP client for sending requests
	Endpoint   string         // Endpoint base URL for the Graph API
}

// Option configures the Teams client.
type Option func(*Client)

// WithHeaders adds custom headers to the request.
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		hc := c.HttpClient.(*utils.HttpClient)
		if hc.Headers == nil {
			hc.Headers = make(map[string]string)
		}
		maps.Copy(hc.Headers, headers)
	}
}

// WithInsecureTLS disables TLS certificate verification.
func WithInsecureTLS(skipInsecure bool) Option {
	return func(c *Client) {
		hc := c.HttpClient.(*utils.HttpClient)
		hc.SkipInsecure = skipInsecure
	}
}

// WithEndpointBase overrides the default Graph API base URL.
func WithEndpointBase(endpoint string) Option {
	return func(c *Client) {
		c.Endpoint = endpoint
	}
}

// WithTimeout sets the per-request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		hc := c.HttpClient.(*utils.HttpClient)
		hc.Timeout = timeout
	}
}

// New returns a new Client with optional configuration.
func New(opts ...Option) *Client {
	c := &Client{
		HttpClient: utils.NewHttpClient(make(map[string]string), false),
		Endpoint:   defaultEndpointBase,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewWithToken returns a Teams client configured with a bearer token.
func NewWithToken(token string, opts ...Option) *Client {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
	return New(append([]Option{WithHeaders(headers)}, opts...)...)
}

// SendChannel sends a message to a Teams channel via Microsoft Graph API.
//
// teamID: ID of the team.
// channelID: ID of the channel inside the team.
func (c *Client) SendChannel(ctx context.Context, teamID, channelID string, msg Message) (*Response, error) {
	base := c.Endpoint
	if base == "" {
		base = defaultEndpointBase
	}
	endpoint := fmt.Sprintf("%s/teams/%s/channels/%s/messages", base, teamID, channelID)
	return c.send(ctx, endpoint, msg)
}

// send marshals the message and handles the HTTP request/response.
func (c *Client) send(ctx context.Context, url string, msg Message) (*Response, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, utils.Wrap(utils.ErrorPermanent, "msteamsgraph marshal", err)
	}

	resp, err := c.HttpClient.DoRequest(ctx, "POST", url, data)
	if err != nil {
		return nil, utils.Wrap(utils.ErrorTransient, "msteamsgraph request", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	var parsed Response
	if err := json.NewDecoder(io.LimitReader(resp.Body, utils.MaxResponseBody)).Decode(&parsed); err != nil {
		return nil, utils.Wrap(utils.ErrorPermanent, "msteamsgraph decode", err)
	}

	if resp.StatusCode >= 400 {
		return &parsed, utils.Wrap(utils.KindFromStatus(resp.StatusCode), "msteamsgraph http status", fmt.Errorf("%s", parsed.ErrorOrMessage()))
	}

	return &parsed, nil
}

// ErrorOrMessage returns a formatted error string from the Graph API response.
func (r *Response) ErrorOrMessage() string {
	if r.Error != nil {
		return fmt.Sprintf("%s: %s", r.Error.Code, r.Error.Message)
	}
	return r.Message
}
