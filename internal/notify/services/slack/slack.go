package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"heartbeats/internal/notify/utils"
	"net/http"
)

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
func New(headers map[string]string, skipTLS bool) *Client {
	return &Client{
		HttpClient: utils.NewHttpClient(headers, skipTLS),
	}
}

// Send sends a Slack message using the Slack client.
func (c *Client) Send(ctx context.Context, slackMessage Slack) (*Response, error) {
	webhookURL := "https://slack.com/api/chat.postMessage"

	data, err := json.Marshal(slackMessage)
	if err != nil {
		return nil, fmt.Errorf("error marshalling Slack message: %v", err)
	}

	resp, err := c.HttpClient.DoRequest(ctx, "POST", webhookURL, data)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if !response.Ok {
		return &response, fmt.Errorf("Slack API error: %s", response.Error)
	}

	return &response, nil
}
