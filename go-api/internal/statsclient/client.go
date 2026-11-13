// Package statsclient calls the Node stats API.
package statsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Stats mirrors the Node API response of POST /api/v1/stats.
type Stats struct {
	Max         float64 `json:"max"`
	Min         float64 `json:"min"`
	Average     float64 `json:"average"`
	Sum         float64 `json:"sum"`
	AnyDiagonal bool    `json:"anyDiagonal"`
	Diagonal    []bool  `json:"diagonal"`
}

// Client talks to the Node stats API at baseURL with a 5 s timeout.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a Client for baseURL (e.g. http://node-api:4000).
func New(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 5 * time.Second}}
}

// Compute posts the matrices to /api/v1/stats, forwarding the caller's Authorization header.
// Any transport error or non-200 status is returned as an error.
func (c *Client) Compute(ctx context.Context, authorization string, matrices ...[][]float64) (*Stats, error) {
	body, err := json.Marshal(map[string]any{"matrices": matrices})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/stats", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stats api returned %d", resp.StatusCode)
	}
	var s Stats
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("decode stats: %w", err)
	}
	return &s, nil
}

// Ping returns nil when the stats API answers GET /health with 200.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("stats api health returned %d", resp.StatusCode)
	}
	return nil
}
