package shodan

import (
	"net/http"
	"sync"
	"time"
)

// Package shodan provides a rate-limited HTTP client for the Shodan API.
// The rate limiter ensures compliance with Shodan's API rate limits by
// restricting requests to a configurable rate (defaults to 2 per second).

// RateLimitedHTTPClient wraps an HTTP client with rate limiting.
// It ensures that no more than the specified number of HTTP requests
// are made per second, which is important for compliance with Shodan's
// API rate limits. The client is thread-safe and can be used concurrently.
type RateLimitedHTTPClient struct {
	client      *http.Client // The underlying HTTP client
	rateLimiter *time.Ticker // Ticker that controls the rate limiting (1 tick per second)
	mu          sync.Mutex   // Mutex for thread-safe operations
}

// NewRateLimitedHTTPClient creates a new rate-limited HTTP client
// that allows at most the specified number of requests per second.
// This is designed to comply with Shodan's API rate limiting requirements.
//
// The rate limiter uses a ticker that fires at the specified interval,
// ensuring that requests are properly spaced out to avoid hitting rate limits.
//
// Parameters:
//   - client: The underlying HTTP client to wrap with rate limiting
//   - requestsPerSecond: Number of requests allowed per second (defaults to 1)
//
// Returns:
//   - A new RateLimitedHTTPClient instance
func NewRateLimitedHTTPClient(client *http.Client, requestsPerSecond int64) *RateLimitedHTTPClient {
	// Ensure minimum rate limit of 1 request per second
	if requestsPerSecond < 1 {
		requestsPerSecond = 1
	}

	// Calculate ticker interval based on requests per second
	interval := time.Duration(1000/requestsPerSecond) * time.Millisecond

	return &RateLimitedHTTPClient{
		client:      client,
		rateLimiter: time.NewTicker(interval),
	}
}

// Do executes an HTTP request with rate limiting
func (r *RateLimitedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Wait for the next tick to ensure rate limiting
	// This ensures we don't exceed 1 request per second
	<-r.rateLimiter.C

	// Execute the request using the underlying client
	return r.client.Do(req)
}

// Close cleans up the rate limiter resources.
// This method stops the ticker and releases associated resources.
// It should be called when the client is no longer needed to
// prevent resource leaks.
//
// The method is thread-safe and can be called concurrently.
func (r *RateLimitedHTTPClient) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.rateLimiter != nil {
		r.rateLimiter.Stop()
	}
}
