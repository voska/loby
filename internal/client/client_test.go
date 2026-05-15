package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voska/loby/internal/errfmt"
)

func newServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := New("test_key", WithBaseURL(ts.URL), WithRetry(2, 5*time.Millisecond))
	return c, ts
}

func TestDo_GET_DecodesBody(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %s, want %s", got, want)
		}
		if u, _, _ := r.BasicAuth(); u != "test_key" {
			t.Fatalf("basic-auth user = %q", u)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Fatal("missing user-agent")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "adr_123", "address_line1": "185 Berry St"})
	})

	out := map[string]any{}
	resp, err := c.Do(context.Background(), &Request{Method: http.MethodGet, Path: "/addresses/adr_123", Out: &out})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if out["id"] != "adr_123" {
		t.Fatalf("decoded body wrong: %v", out)
	}
}

func TestDo_POST_SendsBodyAndIdempotency(t *testing.T) {
	var gotBody map[string]any
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("ctype = %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Idempotency-Key") == "" {
			t.Fatal("expected Idempotency-Key")
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "adr_new"})
	})

	out := map[string]any{}
	_, err := c.Do(context.Background(), &Request{
		Method:         http.MethodPost,
		Path:           "/addresses",
		Body:           map[string]string{"line1": "x"},
		Out:            &out,
		IdempotencyKey: "loby-abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["line1"] != "x" {
		t.Fatalf("body = %v", gotBody)
	}
}

func TestDo_ErrorEnvelope(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message":     "Your API key is not valid.",
				"status_code": 401,
				"code":        "unauthorized",
			},
		})
	})

	_, err := c.Do(context.Background(), &Request{Method: http.MethodGet, Path: "/addresses"})
	if err == nil {
		t.Fatal("expected error")
	}
	ae := AsAPIError(err)
	if ae == nil {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if ae.StatusCode != 401 || ae.Code != "unauthorized" {
		t.Fatalf("unexpected: %+v", ae)
	}
	if got := errfmt.ExitCodeOf(err); got != errfmt.AuthRequired {
		t.Fatalf("exit code = %d, want %d", got, errfmt.AuthRequired)
	}
}

func TestDo_RetriesOn429ThenSucceeds(t *testing.T) {
	var hits int32
	c, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"error":{"message":"slow down","status_code":429}}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	})

	out := map[string]any{}
	resp, err := c.Do(context.Background(), &Request{Method: http.MethodGet, Path: "/x", Out: &out})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != 200 || out["id"] != "ok" {
		t.Fatalf("did not recover: %v / %v", resp.StatusCode, out)
	}
	if atomic.LoadInt32(&hits) != 2 {
		t.Fatalf("hits = %d, want 2", hits)
	}
}

func TestDo_RetryExhaustionReturns429(t *testing.T) {
	var hits int32
	c, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error":{"message":"nope","status_code":429}}`))
	})
	_, err := c.Do(context.Background(), &Request{Method: http.MethodGet, Path: "/x"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := errfmt.ExitCodeOf(err); got != errfmt.RateLimited {
		t.Fatalf("exit code = %d", got)
	}
	if atomic.LoadInt32(&hits) != 3 { // 1 + 2 retries
		t.Fatalf("hits = %d, want 3", hits)
	}
}

func TestDo_Replayed(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Idempotent-Replayed", "true")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"x"}`))
	})
	resp, err := c.Do(context.Background(), &Request{Method: http.MethodPost, Path: "/x", IdempotencyKey: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Replayed {
		t.Fatal("expected Replayed=true")
	}
}

func TestDo_ContextCanceled(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(200)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err := c.Do(ctx, &Request{Method: http.MethodGet, Path: "/x"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIdempotencyKey_Stable(t *testing.T) {
	k1, err := IdempotencyKey("postcards.create", map[string]string{"to": "x"}, map[string]any{"front": "a"}, true)
	if err != nil {
		t.Fatal(err)
	}
	k2, _ := IdempotencyKey("postcards.create", map[string]string{"to": "x"}, map[string]any{"front": "a"}, true)
	if k1 != k2 {
		t.Fatalf("deterministic key changed: %q vs %q", k1, k2)
	}
	if !strings.HasPrefix(k1, "loby-") {
		t.Fatalf("prefix wrong: %q", k1)
	}
}

func TestIdempotencyKey_DifferentInputs(t *testing.T) {
	k1, _ := IdempotencyKey("p", map[string]string{"a": "1"}, nil, true)
	k2, _ := IdempotencyKey("p", map[string]string{"a": "2"}, nil, true)
	if k1 == k2 {
		t.Fatal("expected different keys for different flag values")
	}
}
