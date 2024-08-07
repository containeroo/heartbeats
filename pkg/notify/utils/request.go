package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

// HttpClient represents a custom HTTP client with configurable headers and SSL settings.
type HttpClient struct {
	BaseURL      string            // BaseURL can be set to prefix all request URLs.
	Headers      map[string]string // Headers to be added to each request.
	SkipInsecure bool              // SkipInsecure indicates whether to skip SSL verification.
}

// NewHttpClient creates a new HttpClient with the provided headers and SSL configuration.
//
// Parameters:
//   - headers: HTTP headers to be added to each request.
//   - skipInsecure: Indicates whether to skip SSL verification.
//
// Returns:
//   - *HttpClient: A new instance of HttpClient.
func NewHttpClient(headers map[string]string, skipInsecure bool) *HttpClient {
	return &HttpClient{
		Headers:      headers,
		SkipInsecure: skipInsecure,
	}
}

// DoRequest performs an HTTP request with the specified method, URL, and body.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the HTTP request.
//   - method: The HTTP method to use (e.g., GET, POST).
//   - url: The URL to send the request to.
//   - body: The request body to send.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func (hc *HttpClient) DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	client := hc.createHTTPClient()

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating %s request. %w", method, err)
	}

	// Set custom headers for the request
	for key, value := range hc.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing %s request. %w", method, err)
	}

	return resp, nil
}

// createHTTPClient initializes a new HTTP client with a custom transport configuration.
//
// Returns:
//   - *http.Client: The configured HTTP client.
func (hc *HttpClient) createHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: hc.SkipInsecure},
	}

	return &http.Client{Transport: tr}
}
