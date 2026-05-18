package shodan

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestRateLimiter_FloorIsOneSecond pins the documented contract:
// NewRateLimitedHTTPClient silently raises any interval < 1 to 1. Tests
// that need faster behaviour construct the struct directly (see
// newDirectClient in client_test.go).
func TestRateLimiter_FloorIsOneSecond(t *testing.T) {
	r := NewRateLimitedHTTPClient(&http.Client{}, 0)
	if r.requestInterval != 1 {
		t.Errorf("interval=0 should be raised to 1 (silent floor), got %d", r.requestInterval)
	}
	r = NewRateLimitedHTTPClient(&http.Client{}, -5)
	if r.requestInterval != 1 {
		t.Errorf("negative interval should be raised to 1, got %d", r.requestInterval)
	}
	r = NewRateLimitedHTTPClient(&http.Client{}, 5)
	if r.requestInterval != 5 {
		t.Errorf("explicit interval=5 should pass through, got %d", r.requestInterval)
	}
}

// 0.1.17 regression — 429 must trigger retry-with-backoff. This test
// returns 429 twice then 200; the client should make 3 server hits and
// surface the final 200 without error.
func TestRateLimiter_429RetryThenSuccess(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"error": "rate limited"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `ok`)
	}))
	defer srv.Close()

	rl := &RateLimitedHTTPClient{client: srv.Client(), requestInterval: 0}
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rl.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("final status = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&hits); got != 3 {
		t.Errorf("server hits = %d, want 3 (2 retries + 1 success)", got)
	}
}

// Persistent 429 — after rateLimit429MaxRetries the client returns the 429
// to the caller for surfacing. Total server hits = MaxRetries + 1 (initial).
func TestRateLimiter_429ExhaustRetries(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	rl := &RateLimitedHTTPClient{client: srv.Client(), requestInterval: 0}
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := rl.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429 (surfaced after retries exhausted)", resp.StatusCode)
	}
	wantHits := int32(rateLimit429MaxRetries + 1)
	if got := atomic.LoadInt32(&hits); got != wantHits {
		t.Errorf("server hits = %d, want %d", got, wantHits)
	}
}

// 0.1.17 added body buffering so PUT/POST/DELETE bodies replay correctly
// on 429 retry. Without buffering, the second attempt sends an empty
// body (io.Reader is single-shot) — silently corrupts state-mutating
// calls.
func TestRateLimiter_BodyReplayedOnRetry(t *testing.T) {
	var hits int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rl := &RateLimitedHTTPClient{client: srv.Client(), requestInterval: 0}
	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader(`{"key":"value"}`))
	resp, err := rl.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if len(bodies) != 2 {
		t.Fatalf("expected 2 server hits, got %d (bodies: %v)", len(bodies), bodies)
	}
	if bodies[0] != bodies[1] {
		t.Errorf("body not replayed on retry — first=%q second=%q (0.1.17 regression)", bodies[0], bodies[1])
	}
	if bodies[1] != `{"key":"value"}` {
		t.Errorf("body corrupted on retry: %q", bodies[1])
	}
}
