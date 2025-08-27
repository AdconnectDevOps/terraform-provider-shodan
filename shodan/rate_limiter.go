package shodan

import (
	"net/http"
	"sync"
	"time"
)

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

// Do executes an HTTP request with rate limiting
func (r *RateLimitedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Calculate minimum time between requests based on interval
	minInterval := time.Duration(r.requestInterval) * time.Second

	// If we've made a request before, ensure proper spacing
	if !r.lastRequest.IsZero() {
		timeSinceLast := time.Since(r.lastRequest)
		if timeSinceLast < minInterval {
			// Wait for the remaining time to maintain interval
			waitTime := minInterval - timeSinceLast
			time.Sleep(waitTime)
		}
	}

	// Update last request time
	r.lastRequest = time.Now()

	// Execute the request using the underlying client
	return r.client.Do(req)
}

// Close cleans up the rate limiter resources.
// This method is provided for interface compatibility but doesn't need
// to do anything in the current implementation.
//
// The method is thread-safe and can be called concurrently.
func (r *RateLimitedHTTPClient) Close() {
	// No resources to clean up in the current implementation
}
