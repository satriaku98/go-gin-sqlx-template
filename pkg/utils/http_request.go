package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	// Ambil tracer
	tracer := otel.Tracer("utils/http_request")

	// Mulai span baru untuk HTTP client
	ctx, span := tracer.Start(ctx, "HTTP "+config.Method,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

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
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to marshal request body")
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewBuffer(jsonBody)
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bodyReader)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Inject trace context ke header HTTP
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Set headers dari config
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	// Build curl command
	curlCmd := buildCurlCommand(req, config.Body)

	// Set attributes
	span.SetAttributes(
		attribute.String("http.method", config.Method),
		attribute.String("http.url", config.URL),
		attribute.String("http.curl", curlCmd),
	)

	// Create client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "request failed")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Baca response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Set status code di span
	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)

	return &HttpResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

func buildCurlCommand(req *http.Request, body any) string {
	var b strings.Builder

	b.WriteString("curl -X ")
	b.WriteString(req.Method)

	// Headers
	for k, values := range req.Header {
		for _, v := range values {
			b.WriteString(` -H `)
			b.WriteString(strconv.Quote(fmt.Sprintf("%s: %s", k, v)))
		}
	}

	// Body
	if body != nil {
		switch v := body.(type) {
		case string:
			b.WriteString(" --data ")
			b.WriteString(strconv.Quote(v))
		case []byte:
			b.WriteString(" --data ")
			b.WriteString(strconv.Quote(string(v)))
		default:
			if jsonBody, err := json.Marshal(v); err == nil {
				b.WriteString(" --data ")
				b.WriteString(strconv.Quote(string(jsonBody)))
			}
		}
	}

	// URL (paling akhir)
	b.WriteString(" ")
	b.WriteString(strconv.Quote(req.URL.String()))

	return b.String()
}
