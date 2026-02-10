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
	BaseURL     string
	BearerToken string
	HTTPClient  *http.Client
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
	ToolIDs     []string `json:"tool_ids,omitempty"`
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

// ServerUpdate represents the request body for PUT /servers/{id}.
type ServerUpdate struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ToolIDs     []string `json:"tool_ids,omitempty"`
}

// UpdateServer calls PUT /servers/{id}.
func (c *Client) UpdateServer(ctx context.Context, id string, req ServerUpdate) (*Server, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/servers/"+url.PathEscape(id), req)
	if err != nil {
		return nil, err
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

// --- Gateway types and methods ---

// GatewayHealthCheck represents the health check configuration for a gateway.
type GatewayHealthCheck struct {
	URL      string `json:"url,omitempty"`
	Interval int    `json:"interval,omitempty"`
	Timeout  int    `json:"timeout,omitempty"`
	Retries  int    `json:"retries,omitempty"`
}

// GatewayCreate represents the request body for POST /gateways.
type GatewayCreate struct {
	Name               string                 `json:"name"`
	URL                string                 `json:"url"`
	Description        string                 `json:"description,omitempty"`
	Transport          string                 `json:"transport,omitempty"`
	Capabilities       map[string]interface{} `json:"capabilities,omitempty"`
	HealthCheck        *GatewayHealthCheck    `json:"health_check,omitempty"`
	IsActive           bool                   `json:"is_active"`
	Tags               []string               `json:"tags,omitempty"`
	PassthroughHeaders []string               `json:"passthrough_headers,omitempty"`
	AuthType           string                 `json:"auth_type,omitempty"`
	AuthValue          string                 `json:"auth_value,omitempty"`
}

// GatewayUpdate represents the request body for PUT /gateways/{id}.
type GatewayUpdate struct {
	Name               string                 `json:"name,omitempty"`
	URL                string                 `json:"url,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Transport          string                 `json:"transport,omitempty"`
	Capabilities       map[string]interface{} `json:"capabilities,omitempty"`
	HealthCheck        *GatewayHealthCheck    `json:"health_check,omitempty"`
	IsActive           *bool                  `json:"is_active,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	PassthroughHeaders []string               `json:"passthrough_headers,omitempty"`
	AuthType           string                 `json:"auth_type,omitempty"`
	AuthValue          string                 `json:"auth_value,omitempty"`
}

// Gateway represents a gateway returned by the API.
type Gateway struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	URL                string                 `json:"url"`
	Description        string                 `json:"description,omitempty"`
	Transport          string                 `json:"transport,omitempty"`
	Capabilities       map[string]interface{} `json:"capabilities,omitempty"`
	HealthCheck        *GatewayHealthCheck    `json:"health_check,omitempty"`
	IsActive           bool                   `json:"is_active"`
	Tags               []string               `json:"tags,omitempty"`
	PassthroughHeaders []string               `json:"passthrough_headers,omitempty"`
	AuthType           string                 `json:"auth_type,omitempty"`
	AuthValue          string                 `json:"auth_value,omitempty"`
	CreatedAt          string                 `json:"created_at,omitempty"`
	UpdatedAt          string                 `json:"updated_at,omitempty"`
}

// ListGateways calls GET /gateways.
func (c *Client) ListGateways(ctx context.Context, includeInactive bool) ([]Gateway, error) {
	body, statusCode, err := c.doRequestWithQuery(ctx, http.MethodGet, "/gateways", map[string]string{
		"include_inactive": fmt.Sprintf("%t", includeInactive),
	}, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var gateways []Gateway
	if err := json.Unmarshal(body, &gateways); err != nil {
		return nil, fmt.Errorf("decoding gateways response: %w", err)
	}
	return gateways, nil
}

// CreateGateway calls POST /gateways.
func (c *Client) CreateGateway(ctx context.Context, req GatewayCreate) (*Gateway, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/gateways", req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var gateway Gateway
	if err := json.Unmarshal(body, &gateway); err != nil {
		return nil, fmt.Errorf("decoding gateway response: %w", err)
	}
	return &gateway, nil
}

// GetGateway calls GET /gateways/{id}.
func (c *Client) GetGateway(ctx context.Context, id string) (*Gateway, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/gateways/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var gateway Gateway
	if err := json.Unmarshal(body, &gateway); err != nil {
		return nil, fmt.Errorf("decoding gateway response: %w", err)
	}
	return &gateway, nil
}

// UpdateGateway calls PUT /gateways/{id}.
func (c *Client) UpdateGateway(ctx context.Context, id string, req GatewayUpdate) (*Gateway, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/gateways/"+url.PathEscape(id), req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var gateway Gateway
	if err := json.Unmarshal(body, &gateway); err != nil {
		return nil, fmt.Errorf("decoding gateway response: %w", err)
	}
	return &gateway, nil
}

// DeleteGateway calls DELETE /gateways/{id}.
func (c *Client) DeleteGateway(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/gateways/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
	return nil
}

// --- Tool types and methods ---

// ToolCreate represents the tool fields for creation.
type ToolCreate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// CreateToolRequest represents the request body for POST /tools.
type CreateToolRequest struct {
	Tool       ToolCreate `json:"tool"`
	Visibility string     `json:"visibility,omitempty"`
	TeamID     string     `json:"team_id,omitempty"`
}

// ToolUpdate represents the request body for PUT /tools/{id}.
type ToolUpdate struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// Tool represents a tool returned by the API.
type Tool struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsActive    bool                   `json:"is_active"`
	GatewayID   string                 `json:"gateway_id,omitempty"`
	Visibility  string                 `json:"visibility,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAt   string                 `json:"updated_at,omitempty"`
}

// ListTools calls GET /tools.
func (c *Client) ListTools(ctx context.Context, includeInactive bool) ([]Tool, error) {
	body, statusCode, err := c.doRequestWithQuery(ctx, http.MethodGet, "/tools", map[string]string{
		"include_inactive": fmt.Sprintf("%t", includeInactive),
	}, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var tools []Tool
	if err := json.Unmarshal(body, &tools); err != nil {
		return nil, fmt.Errorf("decoding tools response: %w", err)
	}
	return tools, nil
}

// CreateTool calls POST /tools.
func (c *Client) CreateTool(ctx context.Context, req CreateToolRequest) (*Tool, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/tools", req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var tool Tool
	if err := json.Unmarshal(body, &tool); err != nil {
		return nil, fmt.Errorf("decoding tool response: %w", err)
	}
	return &tool, nil
}

// GetTool calls GET /tools/{id}.
func (c *Client) GetTool(ctx context.Context, id string) (*Tool, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/tools/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var tool Tool
	if err := json.Unmarshal(body, &tool); err != nil {
		return nil, fmt.Errorf("decoding tool response: %w", err)
	}
	return &tool, nil
}

// UpdateTool calls PUT /tools/{id}.
func (c *Client) UpdateTool(ctx context.Context, id string, req ToolUpdate) (*Tool, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/tools/"+url.PathEscape(id), req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var tool Tool
	if err := json.Unmarshal(body, &tool); err != nil {
		return nil, fmt.Errorf("decoding tool response: %w", err)
	}
	return &tool, nil
}

// DeleteTool calls DELETE /tools/{id}.
func (c *Client) DeleteTool(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/tools/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
	return nil
}

// --- Resource types and methods ---

// ResourceCreate represents the resource fields for creation.
type ResourceCreate struct {
	URI         string   `json:"uri"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	MimeType    string   `json:"mimeType,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CreateResourceRequest represents the request body for POST /resources.
type CreateResourceRequest struct {
	Resource   ResourceCreate `json:"resource"`
	Visibility string         `json:"visibility,omitempty"`
	TeamID     string         `json:"team_id,omitempty"`
}

// ResourceUpdate represents the request body for PUT /resources/{id}.
type ResourceUpdate struct {
	URI         string   `json:"uri,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	MimeType    string   `json:"mimeType,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Resource represents a resource returned by the API.
type Resource struct {
	ID          string   `json:"id"`
	URI         string   `json:"uri"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	MimeType    string   `json:"mimeType,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsActive    bool     `json:"is_active"`
	Visibility  string   `json:"visibility,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// ListResources calls GET /resources.
func (c *Client) ListResources(ctx context.Context, includeInactive bool) ([]Resource, error) {
	body, statusCode, err := c.doRequestWithQuery(ctx, http.MethodGet, "/resources", map[string]string{
		"include_inactive": fmt.Sprintf("%t", includeInactive),
	}, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var resources []Resource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, fmt.Errorf("decoding resources response: %w", err)
	}
	return resources, nil
}

// CreateResource calls POST /resources.
func (c *Client) CreateResource(ctx context.Context, req CreateResourceRequest) (*Resource, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/resources", req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var resource Resource
	if err := json.Unmarshal(body, &resource); err != nil {
		return nil, fmt.Errorf("decoding resource response: %w", err)
	}
	return &resource, nil
}

// GetResource calls GET /resources/{id}/info.
func (c *Client) GetResource(ctx context.Context, id string) (*Resource, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/resources/"+url.PathEscape(id)+"/info", nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var resource Resource
	if err := json.Unmarshal(body, &resource); err != nil {
		return nil, fmt.Errorf("decoding resource response: %w", err)
	}
	return &resource, nil
}

// UpdateResource calls PUT /resources/{id}.
func (c *Client) UpdateResource(ctx context.Context, id string, req ResourceUpdate) (*Resource, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/resources/"+url.PathEscape(id), req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var resource Resource
	if err := json.Unmarshal(body, &resource); err != nil {
		return nil, fmt.Errorf("decoding resource response: %w", err)
	}
	return &resource, nil
}

// DeleteResource calls DELETE /resources/{id}.
func (c *Client) DeleteResource(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/resources/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
	return nil
}

// --- Prompt types and methods ---

// PromptArgument represents a single argument in a prompt.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// PromptCreate represents the prompt fields for creation.
type PromptCreate struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
}

// CreatePromptRequest represents the request body for POST /prompts.
type CreatePromptRequest struct {
	Prompt     PromptCreate `json:"prompt"`
	Visibility string       `json:"visibility,omitempty"`
	TeamID     string       `json:"team_id,omitempty"`
}

// PromptUpdate represents the request body for PUT /prompts/{id}.
type PromptUpdate struct {
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
}

// Prompt represents a prompt returned by the API.
type Prompt struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	IsActive    bool             `json:"is_active"`
	Visibility  string           `json:"visibility,omitempty"`
	CreatedAt   string           `json:"created_at,omitempty"`
	UpdatedAt   string           `json:"updated_at,omitempty"`
}

// ListPrompts calls GET /prompts.
func (c *Client) ListPrompts(ctx context.Context, includeInactive bool) ([]Prompt, error) {
	body, statusCode, err := c.doRequestWithQuery(ctx, http.MethodGet, "/prompts", map[string]string{
		"include_inactive": fmt.Sprintf("%t", includeInactive),
	}, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var prompts []Prompt
	if err := json.Unmarshal(body, &prompts); err != nil {
		return nil, fmt.Errorf("decoding prompts response: %w", err)
	}
	return prompts, nil
}

// CreatePrompt calls POST /prompts.
func (c *Client) CreatePrompt(ctx context.Context, req CreatePromptRequest) (*Prompt, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/prompts", req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var prompt Prompt
	if err := json.Unmarshal(body, &prompt); err != nil {
		return nil, fmt.Errorf("decoding prompt response: %w", err)
	}
	return &prompt, nil
}

// GetPrompt calls GET /prompts/{id}.
func (c *Client) GetPrompt(ctx context.Context, id string) (*Prompt, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/prompts/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var prompt Prompt
	if err := json.Unmarshal(body, &prompt); err != nil {
		return nil, fmt.Errorf("decoding prompt response: %w", err)
	}
	return &prompt, nil
}

// UpdatePrompt calls PUT /prompts/{id}.
func (c *Client) UpdatePrompt(ctx context.Context, id string, req PromptUpdate) (*Prompt, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPut, "/prompts/"+url.PathEscape(id), req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var prompt Prompt
	if err := json.Unmarshal(body, &prompt); err != nil {
		return nil, fmt.Errorf("decoding prompt response: %w", err)
	}
	return &prompt, nil
}

// DeletePrompt calls DELETE /prompts/{id}.
func (c *Client) DeletePrompt(ctx context.Context, id string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/prompts/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
	return nil
}

// --- Root types and methods ---

// Root represents a root returned by the API.
type Root struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

// ListRoots calls GET /roots.
func (c *Client) ListRoots(ctx context.Context) ([]Root, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/roots", nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var roots []Root
	if err := json.Unmarshal(body, &roots); err != nil {
		return nil, fmt.Errorf("decoding roots response: %w", err)
	}
	return roots, nil
}

// CreateRoot calls POST /roots.
func (c *Client) CreateRoot(ctx context.Context, req Root) (*Root, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/roots", req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}

	var root Root
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("decoding root response: %w", err)
	}
	return &root, nil
}

// DeleteRoot calls DELETE /roots/{uri}.
func (c *Client) DeleteRoot(ctx context.Context, uri string) error {
	body, statusCode, err := c.doRequest(ctx, http.MethodDelete, "/roots/"+url.PathEscape(uri), nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code %d: %s", statusCode, string(body))
	}
	return nil
}
