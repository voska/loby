package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/voska/loby/internal/errfmt"
)

// APIError is the typed shape of a Lob error response. RetryAfter is populated
// from the Retry-After header for 429 responses.
type APIError struct {
	StatusCode int           `json:"status_code"`
	Code       string        `json:"code,omitempty"`
	Message    string        `json:"message"`
	RequestID  string        `json:"request_id,omitempty"`
	RetryAfter time.Duration `json:"retry_after_ms,omitempty"`
	RawBody    string        `json:"raw,omitempty"`
}

// Error implements error. Format: "lob: <status>: <message> (request <id>)".
func (e *APIError) Error() string {
	parts := []string{fmt.Sprintf("lob %d", e.StatusCode)}
	if e.Code != "" {
		parts = append(parts, e.Code)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	} else if e.RawBody != "" {
		parts = append(parts, e.RawBody)
	}
	out := strings.Join(parts, ": ")
	if e.RequestID != "" {
		out += " (request " + e.RequestID + ")"
	}
	return out
}

// ExitCode maps the Lob status to a loby exit code.
func (e *APIError) ExitCode() int {
	switch e.StatusCode {
	case http.StatusUnauthorized:
		return errfmt.AuthRequired
	case http.StatusNotFound:
		return errfmt.NotFound
	case http.StatusForbidden:
		return errfmt.Forbidden
	case http.StatusPaymentRequired:
		return errfmt.PaymentRequired
	case http.StatusTooManyRequests:
		return errfmt.RateLimited
	case http.StatusRequestTimeout:
		return errfmt.Retryable
	case http.StatusUnprocessableEntity, http.StatusBadRequest:
		return errfmt.UsageError
	default:
		if e.StatusCode >= 500 {
			return errfmt.Retryable
		}
		return errfmt.GeneralError
	}
}

// Transient reports whether the error is worth retrying.
func (e *APIError) Transient() bool {
	switch {
	case e.StatusCode == http.StatusTooManyRequests:
		return true
	case e.StatusCode == http.StatusRequestTimeout:
		return true
	case e.StatusCode >= 500:
		return true
	default:
		return false
	}
}

// errFromResponse parses Lob's error envelope and wraps it with the right exit
// code. Returns nil for 2xx responses.
func errFromResponse(r *Response) error {
	if r == nil || (r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil
	}
	ae := &APIError{StatusCode: r.StatusCode, RequestID: r.RequestID}

	var env struct {
		Error struct {
			Message    string `json:"message"`
			StatusCode int    `json:"status_code"`
			Code       string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(r.RawBody, &env); err == nil && env.Error.Message != "" {
		ae.Message = env.Error.Message
		ae.Code = env.Error.Code
		if env.Error.StatusCode != 0 {
			ae.StatusCode = env.Error.StatusCode
		}
	} else {
		ae.RawBody = string(r.RawBody)
	}

	ae.RetryAfter = retryAfter(r.Headers)
	return errfmt.Wrap(ae.ExitCode(), ae)
}

// retryAfter reads the standard Retry-After header, falling back to Lob's
// X-Rate-Limit-Reset (epoch seconds) when present. Returns zero if neither
// header is parseable.
func retryAfter(h http.Header) time.Duration {
	if v := h.Get("Retry-After"); v != "" {
		if d, ok := parseRetryAfter(v); ok {
			return d
		}
	}
	if v := h.Get("X-Rate-Limit-Reset"); v != "" {
		if secs, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			if d := time.Until(time.Unix(secs, 0)); d > 0 {
				return d
			}
		}
	}
	return 0
}

func parseRetryAfter(v string) (time.Duration, bool) {
	if secs, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d, true
		}
	}
	return 0, false
}

// AsAPIError unwraps err looking for *APIError. Returns nil if not present.
func AsAPIError(err error) *APIError {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}
