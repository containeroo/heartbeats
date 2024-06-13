package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"heartbeats/pkg/notify/utils"
	"net/http"
)

const SLACK_WEBHOOK_URL = "https://slack.com/api/chat.postMessage"

// Slack represents the structure of a Slack message.
type Slack struct {
	Channel     string       `json:"channel"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment represents an attachment in a Slack message.
type Attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
	Title string `json:"title"`
}

// Response represents the structure of a Slack API response.
type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// Client handles the connection and sending of Slack messages.
type Client struct {
	HttpClient *utils.HttpClient
}

// New creates a new Slack client with the given token and TLS setting.
//
// Parameters:
//   - headers: HTTP headers to be used in the requests.
//   - skipTLS: Whether to skip TLS verification.
//
// Returns:
//   - *Client: A new instance of the Client.
func New(headers map[string]string, skipTLS bool) *Client {
	return &Client{
		HttpClient: utils.NewHttpClient(headers, skipTLS),
	}
}

// Send sends a Slack message using the Slack client.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the message sending.
//   - slackMessage: The Slack message to be sent.
//
// Returns:
//   - *Response: The response from the Slack API.
//   - error: An error if sending the message fails.
func (c *Client) Send(ctx context.Context, slackMessage Slack) (*Response, error) {
	data, err := json.Marshal(slackMessage)
	if err != nil {
		return nil, fmt.Errorf("error marshalling Slack message. %w", err)
	}

	resp, err := c.HttpClient.DoRequest(ctx, "POST", SLACK_WEBHOOK_URL, data)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request. %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response. %w", err)
	}

	if !response.Ok {
		return &response, fmt.Errorf("Slack API error: %s", response.Error)
	}

	return &response, nil
}
