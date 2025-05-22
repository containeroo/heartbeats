package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

// HTTPDoer defines the interface required to perform an HTTP request.
//
// It allows consumers to inject custom HTTP clients for testability and flexibility.
type HTTPDoer interface {
	// DoRequest sends an HTTP request using the given parameters.
	//
	// Parameters:
	//   - ctx: A context controlling request timeout and cancellation.
	//   - method: The HTTP method to use (e.g. "GET", "POST").
	//   - url: The target URL of the HTTP request.
	//   - body: The request payload, typically JSON-encoded.
	//
	// Returns:
	//   - *http.Response: The raw HTTP response object.
	//   - error: Any error encountered during request creation or execution.
	DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error)
}

// HttpClient implements the Doer interface with support for custom headers and TLS options.
type HttpClient struct {
	Headers      map[string]string // Headers are added to each outbound HTTP request.
	SkipInsecure bool              // SkipInsecure disables TLS certificate validation when true.
}

// NewHttpClient creates a configured HTTP client for issuing requests.
//
// Parameters:
//   - headers: A map of key-value headers to attach to each request.
//   - skipInsecure: If true, TLS verification is disabled.
//
// Returns:
//   - *HttpClient: A client ready to send requests via DoRequest.
func NewHttpClient(headers map[string]string, skipInsecure bool) *HttpClient {
	return &HttpClient{
		Headers:      headers,
		SkipInsecure: skipInsecure,
	}
}

// DoRequest builds and sends an HTTP request with the given payload and configuration.
//
// Parameters:
//   - ctx: The request-scoped context for cancellation and timeout.
//   - method: The HTTP method to use (e.g. "POST").
//   - url: The request destination.
//   - body: The payload to include in the request body.
//
// Returns:
//   - *http.Response: The HTTP response from the remote server.
//   - error: An error if the request fails to be built or sent.
func (hc *HttpClient) DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	client := hc.createHTTPClient()

	// Build the HTTP request using the provided context.
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", method, err)
	}

	// Add all configured headers to the request.
	for key, value := range hc.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request and return the result.
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing %s request: %w", method, err)
	}

	return resp, nil
}

// createHTTPClient returns an *http.Client configured with TLS and proxy settings.
//
// Returns:
//   - *http.Client: The initialized HTTP client for internal use.
func (hc *HttpClient) createHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: hc.SkipInsecure,
		},
	}

	return &http.Client{
		Transport: transport,
	}
}
