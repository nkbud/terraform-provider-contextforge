// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is the HTTP client for the ContextForge MCP Gateway API.
type Client struct {
	BaseURL    string
	BearerToken string
	HTTPClient *http.Client
}

// NewClient creates a new ContextForge API client.
func NewClient(baseURL, bearerToken string) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		BearerToken: bearerToken,
		HTTPClient:  &http.Client{},
	}
}

// doRequest executes an HTTP request with authentication and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	return c.doRequestWithQuery(ctx, method, path, nil, body)
}

// doRequestWithQuery executes an HTTP request with optional query parameters.
func (c *Client) doRequestWithQuery(ctx context.Context, method, reqPath string, query map[string]string, body interface{}) ([]byte, int, error) {
	reqURL, err := url.JoinPath(c.BaseURL, reqPath)
	if err != nil {
		return nil, 0, fmt.Errorf("building request URL: %w", err)
	}

	if len(query) > 0 {
		parsedURL, err := url.Parse(reqURL)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing request URL: %w", err)
		}
		q := parsedURL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		reqURL = parsedURL.String()
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	if c.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.BearerToken)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// HealthResponse represents the response from GET /health.
type HealthResponse struct {
	Status string `json:"status"`
}

// GetHealth calls GET /health (no auth required).
func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var result HealthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding health response: %w", err)
	}
	return &result, nil
}

// ServerConfig represents the server configuration in create/update requests.
type ServerConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CreateServerRequest represents the request body for POST /servers.
type CreateServerRequest struct {
	Server     ServerConfig `json:"server"`
	Visibility string       `json:"visibility,omitempty"`
}

// Server represents a server returned by the API.
type Server struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
	Status      string   `json:"status,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// ListServers calls GET /servers.
func (c *Client) ListServers(ctx context.Context, includeInactive bool) ([]Server, error) {
	path := "/servers"
	body, statusCode, err := c.doRequestWithQuery(ctx, http.MethodGet, path, map[string]string{
		"include_inactive": fmt.Sprintf("%t", includeInactive),
	}, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var servers []Server
	if err := json.Unmarshal(body, &servers); err != nil {
		return nil, fmt.Errorf("decoding servers response: %w", err)
	}
	return servers, nil
}

// CreateServer calls POST /servers.
func (c *Client) CreateServer(ctx context.Context, req CreateServerRequest) (*Server, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/servers", req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var server Server
	if err := json.Unmarshal(body, &server); err != nil {
		return nil, fmt.Errorf("decoding server response: %w", err)
	}
	return &server, nil
}

// GetServer calls GET /servers/{id}.
func (c *Client) GetServer(ctx context.Context, id string) (*Server, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/servers/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var server Server
	if err := json.Unmarshal(body, &server); err != nil {
		return nil, fmt.Errorf("decoding server response: %w", err)
	}
	return &server, nil
}

// DeleteServer calls DELETE /servers/{id}.
func (c *Client) DeleteServer(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/servers/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
	return nil
}
