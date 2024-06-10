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
func NewHttpClient(headers map[string]string, skipInsecure bool) *HttpClient {
	return &HttpClient{
		Headers:      headers,
		SkipInsecure: skipInsecure,
	}
}

// DoRequest performs an HTTP request with the specified method, URL, and body.
// It returns the HTTP response or an error in case of failure.
func (hc *HttpClient) DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	client := hc.createHTTPClient()

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %v", method, err)
	}

	// Set custom headers for the request
	for key, value := range hc.Headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}

// createHTTPClient initializes a new HTTP client with a custom transport configuration.
func (hc *HttpClient) createHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: hc.SkipInsecure},
	}

	return &http.Client{Transport: tr}
}
