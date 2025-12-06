package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HttpRequestConfig holds configuration for the HTTP request
type HttpRequestConfig struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    any
	Timeout time.Duration
}

// HttpResponse holds the response data
type HttpResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

const (
	defaultTimeout = 30 * time.Second

	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodPatch  = "PATCH"
	MethodDelete = "DELETE"

	ContentTypeJson = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeText = "text/plain"
	ContentTypeXml  = "application/xml"

	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
)

// SendRequest sends an HTTP request using the provided config and context
func SendRequest(ctx context.Context, config HttpRequestConfig) (*HttpResponse, error) {
	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	// Prepare request body
	var bodyReader io.Reader
	if config.Body != nil {
		switch v := config.Body.(type) {
		case string:
			bodyReader = bytes.NewBufferString(v)
		case []byte:
			bodyReader = bytes.NewBuffer(v)
		default:
			jsonBody, err := json.Marshal(config.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewBuffer(jsonBody)
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	// Create client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HttpResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}
