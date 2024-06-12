package msteams

import (
	"context"
	"encoding/json"
	"fmt"
	"heartbeats/pkg/notify/utils"
	"net/http"
)

// MSTeams represents the structure of an MS Teams message.
type MSTeams struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Client handles the connection and sending of MS Teams messages.
type Client struct {
	HttpClient *utils.HttpClient
}

// New creates a new MS Teams client with the given TLS setting.
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

// Send sends an MS Teams message using the MS Teams client.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the message sending.
//   - message: The MS Teams message to be sent.
//   - webhookURL: The webhook URL to send the message to.
//
// Returns:
//   - string: A success message if the message is sent successfully.
//   - error: An error if sending the message fails.
func (c *Client) Send(ctx context.Context, message MSTeams, webhookURL string) (string, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("error marshalling MS Teams message. %w", err)
	}

	resp, err := c.HttpClient.DoRequest(ctx, "POST", webhookURL, data)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request. %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	return "Message sent successfully", nil
}
