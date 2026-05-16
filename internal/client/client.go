// Package client is loby's HTTP client for the Lob v1 API. It exposes one
// generic Do method that resource packages compose into typed methods; every
// request automatically carries a User-Agent, HTTP Basic auth, a generated
// Idempotency-Key for mutations, structured logging, and rate-limit-aware
// retries.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/version"
)

// DefaultBaseURL is Lob's v1 production endpoint. Overridable via Option for
// testing.
const DefaultBaseURL = "https://api.lob.com/v1"

// Client is safe for concurrent use; construct once per CLI invocation.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	userAgent  string
	logger     *slog.Logger
	maxRetries int
	retryBase  time.Duration
}

// Option configures a Client at construction.
type Option func(*Client)

// WithBaseURL overrides the API base URL (testing, staging).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient injects a custom *http.Client (mocks, custom transports).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(ua string) Option { return func(c *Client) { c.userAgent = ua } }

// WithLogger injects a slog.Logger (default: discard).
func WithLogger(l *slog.Logger) Option { return func(c *Client) { c.logger = l } }

// WithRetry configures the retry budget (default: 3 attempts, 500ms base).
func WithRetry(maxAttempts int, base time.Duration) Option {
	return func(c *Client) { c.maxRetries = maxAttempts; c.retryBase = base }
}

// New constructs a Client with the given API key and optional overrides.
func New(apiKey string, opts ...Option) *Client {
	v := version.Get()
	c := &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		userAgent:  fmt.Sprintf("loby/%s (+https://github.com/voska/loby)", v.Version),
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		maxRetries: 3,
		retryBase:  500 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Request bundles everything a Lob call needs. body and form are mutually
// exclusive; body wins. Out is JSON-unmarshaled if non-nil.
type Request struct {
	Method         string
	Path           string
	Query          url.Values
	Body           any
	Form           url.Values
	Files          []FilePart
	IdempotencyKey string
	Out            any
}

// FilePart represents one file in a multipart upload.
type FilePart struct {
	Field    string
	Filename string
	Reader   io.Reader
}

// Response surfaces useful metadata in addition to the unmarshaled body.
type Response struct {
	StatusCode int
	Headers    http.Header
	RequestID  string
	Replayed   bool
	RawBody    []byte
}

// Do executes req. It applies auth + idempotency + retries and unmarshals
// req.Out when set. Returns the decoded *APIError (wrapped with the right
// exit code) for any non-2xx response.
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	if req.Method == "" {
		return nil, errfmt.Wrap(errfmt.GeneralError, errors.New("client: empty Request.Method"))
	}

	body, contentType, err := buildBody(req)
	if err != nil {
		return nil, errfmt.Wrap(errfmt.GeneralError, err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.backoff(attempt, lastErr)
			c.logger.DebugContext(ctx, "retrying", "attempt", attempt, "delay", delay, "err", lastErr)
			select {
			case <-ctx.Done():
				return nil, errfmt.Wrap(errfmt.Retryable, ctx.Err())
			case <-time.After(delay):
			}
		}

		resp, err := c.send(ctx, req, body, contentType)
		if err != nil {
			lastErr = err
			if !isTransient(err) {
				return nil, err
			}
			continue
		}
		if shouldRetry(resp.StatusCode) {
			lastErr = errFromResponse(resp)
			if attempt < c.maxRetries {
				continue
			}
		}
		return resp, errFromResponse(resp)
	}
	return nil, lastErr
}

func (c *Client) send(ctx context.Context, req *Request, body []byte, contentType string) (*Response, error) {
	u, err := c.url(req.Path, req.Query)
	if err != nil {
		return nil, errfmt.Wrap(errfmt.GeneralError, err)
	}

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	hr, err := http.NewRequestWithContext(ctx, req.Method, u, bodyReader)
	if err != nil {
		return nil, errfmt.Wrap(errfmt.GeneralError, fmt.Errorf("build request: %w", err))
	}
	hr.SetBasicAuth(c.apiKey, "")
	hr.Header.Set("User-Agent", c.userAgent)
	hr.Header.Set("Accept", "application/json")
	if contentType != "" {
		hr.Header.Set("Content-Type", contentType)
	}
	if req.IdempotencyKey != "" && isMutation(req.Method) {
		hr.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}

	c.logger.DebugContext(ctx, "request", "method", req.Method, "url", u, "idem", req.IdempotencyKey)

	httpResp, err := c.httpClient.Do(hr)
	if err != nil {
		return nil, errfmt.Wrap(errfmt.Retryable, fmt.Errorf("http: %w", err))
	}
	defer func() { _ = httpResp.Body.Close() }()

	raw, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errfmt.Wrap(errfmt.Retryable, fmt.Errorf("read body: %w", err))
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		RequestID:  httpResp.Header.Get("X-Rate-Limit-Request-Id"),
		Replayed:   strings.EqualFold(httpResp.Header.Get("Idempotent-Replayed"), "true"),
		RawBody:    raw,
	}
	if resp.RequestID == "" {
		resp.RequestID = httpResp.Header.Get("X-Request-Id")
	}

	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 && req.Out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, req.Out); err != nil {
			return resp, errfmt.Wrap(errfmt.GeneralError, fmt.Errorf("decode response: %w", err))
		}
	}
	return resp, nil
}

// url assembles baseURL + path + encoded query. path may be absolute or relative.
func (c *Client) url(p string, q url.Values) (string, error) {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	full := c.baseURL + p
	if len(q) > 0 {
		full += "?" + q.Encode()
	}
	if _, err := url.Parse(full); err != nil {
		return "", fmt.Errorf("invalid url %q: %w", full, err)
	}
	return full, nil
}

func (c *Client) backoff(attempt int, lastErr error) time.Duration {
	var ae *APIError
	if errors.As(lastErr, &ae) && ae.RetryAfter > 0 {
		return ae.RetryAfter
	}
	return c.retryBase * (1 << (attempt - 1))
}

// shouldRetry mirrors APIError.Transient at the status-code level for the
// retry loop. Kept in client.go so the loop has no reflection cost on the hot
// path.
func shouldRetry(status int) bool {
	switch {
	case status == http.StatusTooManyRequests:
		return true
	case status == http.StatusRequestTimeout:
		return true
	case status >= 500:
		return true
	default:
		return false
	}
}

func isMutation(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isTransient(err error) bool {
	if err == nil {
		return false
	}
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.Transient()
	}
	var coded *errfmt.Coded
	if errors.As(err, &coded) {
		return coded.ExitCode == errfmt.Retryable || coded.ExitCode == errfmt.RateLimited
	}
	return false
}
