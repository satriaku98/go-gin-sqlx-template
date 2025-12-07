package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	var bodyBytes []byte

	if config.Body != nil {
		switch v := config.Body.(type) {
		case string:
			bodyBytes = []byte(v)
		case []byte:
			bodyBytes = v
		default:
			jsonBody, err := json.Marshal(config.Body)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to marshal request body")
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyBytes = jsonBody
		}
	}

	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
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
	curlCmd := buildCurlCommand(req, bodyBytes)

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
		attribute.String("http.response_body", string(respBody)),
	)

	return &HttpResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

func buildCurlCommand(req *http.Request, rawBody []byte) string {
	var b strings.Builder

	b.WriteString("curl -X ")
	b.WriteString(req.Method)
	b.WriteString(" '")
	b.WriteString(buildRedactedURL(req.URL))
	b.WriteString("'")

	// headers
	for name, values := range req.Header {
		for _, v := range values {
			headerVal := v
			if isSensitiveKey(name) {
				headerVal = "***REDACTED***"
			}
			b.WriteString(" \\\n  -H '")
			b.WriteString(name)
			b.WriteString(": ")
			b.WriteString(headerVal)
			b.WriteString("'")
		}
	}

	// body
	if len(rawBody) > 0 {
		contentType := req.Header.Get("Content-Type")
		redBody := redactBody(rawBody, contentType)
		if len(redBody) > 0 {
			b.WriteString(" \\\n  --data '")
			bodyStr := string(redBody)
			// escape single quote biar aman di shell
			bodyStr = strings.ReplaceAll(bodyStr, `'`, `'\''`)
			b.WriteString(bodyStr)
			b.WriteString("'")
		}
	}

	return b.String()
}

func isSensitiveKey(key string) bool {
	switch strings.ToLower(key) {
	case
		"authorization",
		"proxy-authorization",
		"cookie",
		"set-cookie",
		"x-api-key",
		"x-api-token",
		"x-access-token",
		"x-auth-token",
		"apikey",
		"api-key",
		"password",
		"pass",
		"pwd",
		"token",
		"access_token",
		"refresh_token",
		"client_secret",
		"secret":
		return true
	default:
		return false
	}
}

func buildRedactedURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	clone := *u
	q := clone.Query()

	for key := range q {
		if isSensitiveKey(key) {
			q.Set(key, "***REDACTED***")
		}
	}
	clone.RawQuery = q.Encode()

	return clone.String()
}

func redactJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m2 := make(map[string]any, len(t))
		for k, val := range t {
			if isSensitiveKey(k) {
				m2[k] = "***REDACTED***"
			} else {
				m2[k] = redactJSON(val)
			}
		}
		return m2
	case []any:
		for i, val := range t {
			t[i] = redactJSON(val)
		}
		return t
	default:
		return v
	}
}

func redactBody(body []byte, contentType string) []byte {
	if len(body) == 0 {
		return body
	}

	ct := strings.ToLower(contentType)

	// JSON body
	if strings.Contains(ct, "application/json") {
		var v any
		if err := json.Unmarshal(body, &v); err != nil {
			// kalau gagal parsing, lebih aman di-redact penuh
			return []byte("<redacted>")
		}
		v = redactJSON(v)
		out, err := json.Marshal(v)
		if err != nil {
			return []byte("<redacted>")
		}
		return out
	}

	// Form body
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return []byte("<redacted>")
		}
		for key := range vals {
			if isSensitiveKey(key) {
				vals.Set(key, "***REDACTED***")
			}
		}
		return []byte(vals.Encode())
	}

	return body
}
