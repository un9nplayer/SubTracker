// Package api provides the HTTP client for the AgniOps Subdomain Scan API.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	scanEndpoint = "https://app.agniops.in/api/v1/subdomains/scan"
	userAgent    = "SubTracker/1.0.0 (github.com/subtracker)"
)

// ─── Response types ────────────────────────────────────────────────────────────

// SubdomainEntry represents a single discovered subdomain record.
type SubdomainEntry struct {
	Subdomain  string `json:"subdomain"`
	IP         string `json:"ip"`
	Cloudflare bool   `json:"cloudflare"`
}

// Meta holds API quota information returned with every scan.
type Meta struct {
	QuotaRemaining int `json:"quota_remaining"`
	DailyQuota     int `json:"daily_quota"`
}

// ScanResult is the full AgniOps API response body.
type ScanResult struct {
	Status          string           `json:"status"`
	Engine          string           `json:"engine"`
	Domain          string           `json:"domain"`
	SubdomainsCount int              `json:"subdomains_count"`
	Country         string           `json:"country"`
	MostUsedIP      string           `json:"most_used_ip"`
	ScanDate        string           `json:"scan_date"`
	Subdomains      []SubdomainEntry `json:"subdomains"`
	Meta            Meta             `json:"meta"`
}

// ─── Client ────────────────────────────────────────────────────────────────────

// Client is an authenticated HTTP client for the AgniOps API.
type Client struct {
	apiKey string
	http   *http.Client
}

// NewClient creates a new API client with the given key and per-request timeout.
func NewClient(apiKey string, timeout time.Duration) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

// Scan performs a subdomain discovery scan for the given domain and returns
// the structured API response. It returns a descriptive error for all non-2xx
// responses, including auth failures and rate limiting.
func (c *Client) Scan(domain string) (*ScanResult, error) {
	// Build request body
	payload, err := json.Marshal(map[string]string{"domain": domain})
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, scanEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", userAgent)

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error — check your internet connection: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle well-known error codes with helpful messages
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf(
			"authentication failed (HTTP %d) — verify your API key with: subtracker configure",
			resp.StatusCode,
		)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf(
			"rate limit exceeded (HTTP 429) — daily quota exhausted. Check your quota at AgniOps dashboard",
		)
	case http.StatusNotFound:
		return nil, fmt.Errorf("API endpoint not found (HTTP 404) — contact AgniOps support")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, fmt.Errorf("AgniOps server error (HTTP %d) — try again later", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		// Attempt to extract an error message from the response body
		var apiErr struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil {
			msg := apiErr.Message
			if msg == "" {
				msg = apiErr.Error
			}
			if msg != "" {
				return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var result ScanResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &result, nil
}
