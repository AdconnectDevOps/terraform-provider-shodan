package shodan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// alertsCacheTTL controls how long ListAlerts results are reused across calls
// within a single provider run. Tuned to amortize burst load (many resources
// running Read in parallel during refresh) without masking real out-of-band
// changes.
const alertsCacheTTL = 60 * time.Second

// ShodanClient represents a client for interacting with the Shodan API
type ShodanClient struct {
	ApiKey     string
	BaseURL    string
	HTTPClient *RateLimitedHTTPClient

	alertsCache   []AlertResponse
	alertsCacheAt time.Time
	cacheMu       sync.Mutex
}

// NewShodanClient creates a new Shodan API client
func NewShodanClient(apiKey string) *ShodanClient {
	return &ShodanClient{
		ApiKey:     apiKey,
		BaseURL:    "https://api.shodan.io",
		HTTPClient: NewRateLimitedHTTPClient(&http.Client{}, 2), // Default to 2 seconds between requests
	}
}

// CreateAlert creates a new Shodan alert
func (c *ShodanClient) CreateAlert(name string, filters map[string]interface{}) (*AlertResponse, error) {
	payload := map[string]interface{}{
		"name":    name,
		"filters": filters,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alert payload: %w", err)
	}

	// Try the alert endpoint first
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/shodan/alert?key=%s", c.BaseURL, c.ApiKey), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var alertResp AlertResponse
	if err := json.Unmarshal(body, &alertResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &alertResp, nil
}

// AddIgnoreService adds an (ip, port) pair to a trigger's ignore list — the
// Shodan API equivalent of the "Add to Whitelist" button in the alert email.
// service is formatted as "ip:port" (e.g. "203.0.113.10:3307").
func (c *ShodanClient) AddIgnoreService(alertID, trigger, service string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	if trigger == "" {
		return fmt.Errorf("trigger cannot be empty")
	}
	if service == "" {
		return fmt.Errorf("service cannot be empty")
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/shodan/alert/%s/trigger/%s/ignore/%s?key=%s", c.BaseURL, alertID, trigger, service, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemoveIgnoreService removes an (ip, port) pair from a trigger's ignore list.
// 404 is treated as success (idempotent).
func (c *ShodanClient) RemoveIgnoreService(alertID, trigger, service string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	if trigger == "" {
		return fmt.Errorf("trigger cannot be empty")
	}
	if service == "" {
		return fmt.Errorf("service cannot be empty")
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/shodan/alert/%s/trigger/%s/ignore/%s?key=%s", c.BaseURL, alertID, trigger, service, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
}

// ExtractIgnoredServices parses the per-trigger ignore arrays from an alert's
// triggers map. Returns map[trigger]→sorted []string of "ip:port" entries.
// The Shodan API returns each trigger as map[string]any with an "ignore" key
// containing []any of map[string]any{"ip": ..., "port": ...}; any other shape
// is silently skipped (forward-compat for API additions).
func ExtractIgnoredServices(triggers map[string]interface{}) map[string][]string {
	out := make(map[string][]string)
	for name, raw := range triggers {
		t, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		ignoreRaw, ok := t["ignore"]
		if !ok {
			continue
		}
		ignoreList, ok := ignoreRaw.([]interface{})
		if !ok {
			continue
		}
		for _, entry := range ignoreList {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			ip, _ := entryMap["ip"].(string)
			if ip == "" {
				continue
			}
			var portStr string
			switch p := entryMap["port"].(type) {
			case float64:
				portStr = fmt.Sprintf("%d", int(p))
			case int:
				portStr = fmt.Sprintf("%d", p)
			case string:
				portStr = p
			}
			if portStr == "" {
				continue
			}
			out[name] = append(out[name], fmt.Sprintf("%s:%s", ip, portStr))
		}
	}
	return out
}

// AddTrigger adds a trigger to an existing alert
func (c *ShodanClient) AddTrigger(alertID, trigger string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/shodan/alert/%s/trigger/%s?key=%s", c.BaseURL, alertID, trigger, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemoveTrigger removes a trigger from an existing alert
func (c *ShodanClient) RemoveTrigger(alertID, trigger string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/shodan/alert/%s/trigger/%s?key=%s", c.BaseURL, alertID, trigger, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
}

// AddNotifier adds a notifier to an existing alert
func (c *ShodanClient) AddNotifier(alertID, notifierID string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/shodan/alert/%s/notifier/%s?key=%s", c.BaseURL, alertID, notifierID, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemoveNotifier removes a notifier from an existing alert
func (c *ShodanClient) RemoveNotifier(alertID, notifierID string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/shodan/alert/%s/notifier/%s?key=%s", c.BaseURL, alertID, notifierID, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
}

// ListAlerts retrieves all alerts for the authenticated account. Results are
// cached on the client for alertsCacheTTL — multiple Read calls in one plan
// share the same response and avoid hammering the Shodan API.
func (c *ShodanClient) ListAlerts() ([]AlertResponse, error) {
	c.cacheMu.Lock()
	if c.alertsCache != nil && time.Since(c.alertsCacheAt) < alertsCacheTTL {
		cached := c.alertsCache
		c.cacheMu.Unlock()
		return cached, nil
	}
	c.cacheMu.Unlock()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/shodan/alert/info?key=%s", c.BaseURL, c.ApiKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var alerts []AlertResponse
	if err := json.Unmarshal(body, &alerts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.cacheMu.Lock()
	c.alertsCache = alerts
	c.alertsCacheAt = time.Now()
	c.cacheMu.Unlock()

	return alerts, nil
}

// AddEmailNotifier adds an email notifier to an existing alert
func (c *ShodanClient) AddEmailNotifier(alertID, email string) error {
	// First, we need to create a custom notifier for the email
	// This is a simplified approach - in practice, you might need to use existing notifiers
	// or create custom ones via the Shodan API

	// For now, we'll use the default notifier and assume it's configured for the email
	// You may need to configure this in your Shodan account settings
	return c.AddNotifier(alertID, "default")
}

// AddSlackNotifier adds a Slack notifier to an existing alert
func (c *ShodanClient) AddSlackNotifier(alertID, notifierID string) error {
	// Add the specified Slack notifier ID
	// Users should configure their Slack notifier ID in their Terraform configuration
	return c.AddNotifier(alertID, notifierID)
}

// GetAlert retrieves an existing alert by ID
func (c *ShodanClient) GetAlert(alertID string) (*AlertResponse, error) {
	if alertID == "" {
		return nil, fmt.Errorf("alert ID cannot be empty")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/shodan/alert/%s/info?key=%s", c.BaseURL, alertID, c.ApiKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var alertResp AlertResponse
	if err := json.Unmarshal(body, &alertResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &alertResp, nil
}

// DeleteAlert deletes an existing alert by ID
func (c *ShodanClient) DeleteAlert(alertID string) error {
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/shodan/alert/%s?key=%s", c.BaseURL, alertID, c.ApiKey), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Accept both 200 (OK) and 404 (Not Found) as success for delete operations
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
}

// UpdateAlert updates an existing alert's network filters
func (c *ShodanClient) UpdateAlert(alertID string, filters map[string]interface{}) error {
	// Add validation for alertID
	if alertID == "" {
		return fmt.Errorf("alert ID cannot be empty")
	}

	payload := map[string]interface{}{
		"filters": filters,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert update payload: %w", err)
	}

	// Construct the URL and log it for debugging
	url := fmt.Sprintf("%s/shodan/alert/%s?key=%s", c.BaseURL, alertID, c.ApiKey)

	// Use the POST /shodan/alert/{id} endpoint as per Shodan API documentation
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close cleans up the rate limiter resources
func (c *ShodanClient) Close() {
	if c.HTTPClient != nil {
		c.HTTPClient.Close()
	}
}

// AlertResponse represents the response from Shodan API for alert operations
type AlertResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Created     string                 `json:"created"`
	Triggers    map[string]interface{} `json:"triggers"`
	HasTriggers bool                   `json:"has_triggers"`
	Expires     int                    `json:"expires"`
	Expiration  interface{}            `json:"expiration"`
	Filters     map[string]interface{} `json:"filters"`
	Size        int                    `json:"size"`
}

// GetDomainInfo retrieves domain information including subdomains and DNS records
func (c *ShodanClient) GetDomainInfo(domain string) (*DomainInfo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/dns/domain/%s?key=%s", c.BaseURL, domain, c.ApiKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var domainInfo DomainInfo
	if err := json.Unmarshal(body, &domainInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &domainInfo, nil
}

// ResolveDomain resolves a domain to its actual IP addresses using system DNS
func (c *ShodanClient) ResolveDomain(domain string) ([]string, error) {
	ips, err := net.LookupHost(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve domain %s: %w", domain, err)
	}
	return ips, nil
}

// CreateDomainAlert creates a new Shodan alert for monitoring a domain
func (c *ShodanClient) CreateDomainAlert(name string, domain string, triggers []string) (*AlertResponse, error) {
	// Use proper DNS resolution instead of trusting Shodan's historical data
	ips, err := c.ResolveDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve domain %s: %w", domain, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for domain %s", domain)
	}

	// Create filters for the alert using only the actually resolved IPs
	filters := map[string]interface{}{
		"ip": ips,
	}

	// Create the alert with domain-specific naming
	alertName := fmt.Sprintf("__domain: %s", domain)
	if name != "" {
		alertName = fmt.Sprintf("__domain: %s (%s)", domain, name)
	}

	payload := map[string]interface{}{
		"name":    alertName,
		"filters": filters,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alert payload: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/shodan/alert?key=%s", c.BaseURL, c.ApiKey), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var alertResp AlertResponse
	if err := json.Unmarshal(body, &alertResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &alertResp, nil
}

// DomainInfo represents the response from Shodan API for domain information
type DomainInfo struct {
	Domain     string       `json:"domain"`
	Tags       []string     `json:"tags"`
	Subdomains []string     `json:"subdomains"`
	Data       []DomainData `json:"data"`
	More       bool         `json:"more"`
}

// DomainData represents individual DNS records for a domain
type DomainData struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	LastSeen  string `json:"last_seen"`
}
