package targets

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notify/render"
	"github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/templates"
)

// WebhookTarget delivers notifications to a webhook endpoint.
type WebhookTarget struct {
	URL       string
	Headers   map[string]string
	Template  *templates.Template
	TitleTmpl *templates.StringTemplate
	Client    *http.Client
	Logger    *slog.Logger
	Receiver  string
	Vars      map[string]any
}

// NewWebhookTarget constructs a webhook target.
func NewWebhookTarget(
	receiver string,
	url string,
	headers map[string]string,
	tmpl *templates.Template,
	titleTmpl *templates.StringTemplate,
	vars map[string]any,
	logger *slog.Logger,
) *WebhookTarget {
	return &WebhookTarget{
		URL:       url,
		Headers:   headers,
		Template:  tmpl,
		TitleTmpl: titleTmpl,
		Client:    newHTTPClient(),
		Logger:    logger,
		Receiver:  receiver,
		Vars:      vars,
	}
}

// Type returns the target type name.
func (w *WebhookTarget) Type() string { return types.TargetWebhook.String() }

// Send renders and posts a webhook notification.
func (w *WebhookTarget) Send(n types.Payload) error {
	_, err := w.SendResult(n)
	return err
}

// SendResult renders and posts a webhook notification and returns response details.
func (w *WebhookTarget) SendResult(n types.Payload) (types.DeliveryResult, error) {
	if w == nil {
		return types.DeliveryResult{}, errors.New("webhook target is nil")
	}
	data := render.NewRenderData(n, w.Receiver, w.Vars, "")
	subject, err := w.TitleTmpl.Render(data)
	if err != nil {
		return types.DeliveryResult{}, err
	}
	data.Subject = subject
	body, err := w.Template.Render(data)
	if err != nil {
		return types.DeliveryResult{}, fmt.Errorf("render webhook template: %w", err)
	}
	client := w.Client
	if client == nil {
		client = newHTTPClient()
	}
	status, statusCode, respBody, err := postBody(client, w.URL, w.Headers, body, "application/json")
	w.Logger.Debug("Webhook response",
		"event", logging.EventWebhookResponse.String(),
		"receiver", w.Receiver,
		"url", w.URL,
		"status", status,
	)
	if err != nil {
		return types.DeliveryResult{
			Status:     status,
			StatusCode: statusCode,
			Response:   truncateBody(respBody, 2048),
		}, err
	}
	if status != "200 OK" {
		return types.DeliveryResult{
			Status:     status,
			StatusCode: statusCode,
			Response:   truncateBody(respBody, 2048),
		}, fmt.Errorf("webhook delivery failed: %s (%s)", status, respBody)
	}
	return types.DeliveryResult{
		Status:     status,
		StatusCode: statusCode,
		Response:   truncateBody(respBody, 2048),
	}, nil
}

// newHTTPClient creates an HTTP client honoring proxy environment variables.
func newHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
}

// postBody posts a raw body to a URL with optional headers.
func postBody(
	client *http.Client,
	url string,
	headers map[string]string,
	body []byte,
	contentType string,
) (status string, statusCode int, response string, err error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", 0, "", err
	}
	req.Header.Set("Content-Type", contentType+"; charset=utf-8")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close() // nolint:errcheck
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	respText := strings.TrimSpace(string(respBody))
	if resp.StatusCode >= 300 {
		return resp.Status, resp.StatusCode, respText, fmt.Errorf("non-2xx response: %s (%s)", resp.Status, respText)
	}
	return resp.Status, resp.StatusCode, respText, nil
}

// truncateBody truncates a body to a given length.
func truncateBody(body string, limit int) string {
	if limit <= 0 || len(body) <= limit {
		return body
	}
	return body[:limit] + "..."
}
