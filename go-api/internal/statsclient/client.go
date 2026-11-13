// Package statsclient llama a la API de estadísticas en Node.
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

// Stats refleja la respuesta de POST /api/v1/stats.
type Stats struct {
	Max         float64 `json:"max"`
	Min         float64 `json:"min"`
	Average     float64 `json:"average"`
	Sum         float64 `json:"sum"`
	AnyDiagonal bool    `json:"anyDiagonal"`
	Diagonal    []bool  `json:"diagonal"`
}

// Client habla con la API Node en baseURL con timeout de 5 s.
type Client struct {
	baseURL string
	http    *http.Client
}

// New crea un Client para baseURL (p. ej. http://node-api:4000).
func New(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 5 * time.Second}}
}

// Compute envía las matrices a /api/v1/stats reenviando el header Authorization del cliente.
// Cualquier error de transporte o status distinto de 200 se devuelve como error.
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

// Ping devuelve nil si la API responde 200 a GET /health.
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
