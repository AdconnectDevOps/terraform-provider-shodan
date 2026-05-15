package shodan

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// rateLimit429MaxRetries is the maximum number of times Do will retry on a
// 429 response. With exponential backoff (interval × 2^attempt) the total
// wait for an interval of 1s is roughly 1+2+4 = 7s before giving up.
const rateLimit429MaxRetries = 3

// Package shodan provides a rate-limited HTTP client for the Shodan API.
// The rate limiter ensures compliance with Shodan's API rate limits by
// restricting requests to a configurable interval (defaults to 2 seconds between requests).

// RateLimitedHTTPClient wraps an HTTP client with rate limiting.
// It ensures that HTTP requests are spaced at least the specified interval
// apart, which is important for compliance with Shodan's API rate limits.
// The client is thread-safe and can be used concurrently.
type RateLimitedHTTPClient struct {
	client          *http.Client // The underlying HTTP client
	requestInterval int64        // Minimum seconds between requests
	lastRequest     time.Time    // Time of the last request
	mu              sync.Mutex   // Mutex for thread-safe operations
}

// NewRateLimitedHTTPClient creates a new rate-limited HTTP client
// that ensures requests are spaced at least the specified interval apart.
// This is designed to comply with Shodan's API rate limiting requirements.
//
// The rate limiter uses a time-based approach to ensure proper spacing
// between requests based on the specified interval.
//
// Parameters:
//   - client: The underlying HTTP client to wrap with rate limiting
//   - requestIntervalSeconds: Minimum seconds between requests (defaults to 2)
//
// Returns:
//   - A new RateLimitedHTTPClient instance
func NewRateLimitedHTTPClient(client *http.Client, requestIntervalSeconds int64) *RateLimitedHTTPClient {
	// Ensure minimum interval of 1 second between requests
	if requestIntervalSeconds < 1 {
		requestIntervalSeconds = 1
	}

	return &RateLimitedHTTPClient{
		client:          client,
		requestInterval: requestIntervalSeconds,
		lastRequest:     time.Time{}, // Zero time means no previous request
	}
}

// Do executes an HTTP request with rate limiting and 429 retry-with-backoff.
// Shodan's rate limit advertises 1 request per second, but their server-side
// bucket can flag bursts even when the wall-clock spacing is correct. On 429
// we sleep for (interval × 2^attempt) and retry, up to rateLimit429MaxRetries.
func (r *RateLimitedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	minInterval := time.Duration(r.requestInterval) * time.Second

	// Buffer the body so we can replay on retry. GET requests have no body, but
	// PUT/POST/DELETE may — and io.Reader-style bodies are single-shot.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to buffer request body for retry: %w", err)
		}
		req.Body.Close()
	}

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= rateLimit429MaxRetries; attempt++ {
		if !r.lastRequest.IsZero() {
			timeSinceLast := time.Since(r.lastRequest)
			if timeSinceLast < minInterval {
				time.Sleep(minInterval - timeSinceLast)
			}
		}

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		r.lastRequest = time.Now()
		resp, lastErr = r.client.Do(req)
		if lastErr != nil {
			return nil, lastErr
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// 429: drain the body so we can retry the same connection, then backoff.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if attempt == rateLimit429MaxRetries {
			return resp, nil // last attempt — return the 429 to the caller for surfacing
		}

		backoff := minInterval * time.Duration(1<<attempt)
		time.Sleep(backoff)
	}

	return resp, nil
}

// Close cleans up the rate limiter resources.
// This method is provided for interface compatibility but doesn't need
// to do anything in the current implementation.
//
// The method is thread-safe and can be called concurrently.
func (r *RateLimitedHTTPClient) Close() {
	// No resources to clean up in the current implementation
}
