// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "")
	health, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != "ok" {
		t.Errorf("expected status ok, got %s", health.Status)
	}
}

func TestListServers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers" {
			t.Errorf("expected path /servers, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected auth header, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Server{
			{ID: "srv-1", Name: "test-server"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	servers, err := c.ListServers(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].ID != "srv-1" {
		t.Errorf("expected server ID srv-1, got %s", servers[0].ID)
	}
}

func TestCreateServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/servers" {
			t.Errorf("expected path /servers, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req CreateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Server.Name != "my-server" {
			t.Errorf("expected server name my-server, got %s", req.Server.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Server{
			ID:         "srv-new",
			Name:       req.Server.Name,
			Visibility: req.Visibility,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	srv, err := c.CreateServer(context.Background(), CreateServerRequest{
		Server:     ServerConfig{Name: "my-server", Description: "A test server"},
		Visibility: "private",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.ID != "srv-new" {
		t.Errorf("expected server ID srv-new, got %s", srv.ID)
	}
}

func TestGetServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers/srv-1" {
			t.Errorf("expected path /servers/srv-1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Server{ID: "srv-1", Name: "test-server"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	srv, err := c.GetServer(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("expected server, got nil")
	}
	if srv.ID != "srv-1" {
		t.Errorf("expected server ID srv-1, got %s", srv.ID)
	}
}

func TestGetServer_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	srv, err := c.GetServer(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv != nil {
		t.Errorf("expected nil server for 404, got %v", srv)
	}
}

func TestDeleteServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/servers/srv-1" {
			t.Errorf("expected path /servers/srv-1, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteServer(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/servers/srv-1" {
			t.Errorf("expected path /servers/srv-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req ServerUpdate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Name != "updated-server" {
			t.Errorf("expected server name updated-server, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Server{
			ID:          "srv-1",
			Name:        req.Name,
			Description: req.Description,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	srv, err := c.UpdateServer(context.Background(), "srv-1", ServerUpdate{
		Name:        "updated-server",
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.Name != "updated-server" {
		t.Errorf("expected server name updated-server, got %s", srv.Name)
	}
}

// --- Gateway Tests ---

func TestCreateGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/gateways" {
			t.Errorf("expected path /gateways, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req GatewayCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Name != "test-gw" {
			t.Errorf("expected gateway name test-gw, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Gateway{
			ID:          "gw-1",
			Name:        req.Name,
			URL:         req.URL,
			Description: req.Description,
			Transport:   req.Transport,
			IsActive:    req.IsActive,
			Tags:        req.Tags,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	gw, err := c.CreateGateway(context.Background(), GatewayCreate{
		Name:        "test-gw",
		URL:         "https://example.com",
		Description: "Test",
		Transport:   "STREAMABLEHTTP",
		IsActive:    true,
		Tags:        []string{"test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.ID != "gw-1" {
		t.Errorf("expected gateway ID gw-1, got %s", gw.ID)
	}
	if gw.Name != "test-gw" {
		t.Errorf("expected gateway name test-gw, got %s", gw.Name)
	}
}

func TestGetGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gateways/gw-1" {
			t.Errorf("expected path /gateways/gw-1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Gateway{ID: "gw-1", Name: "test-gw"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	gw, err := c.GetGateway(context.Background(), "gw-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw == nil {
		t.Fatal("expected gateway, got nil")
	}
	if gw.ID != "gw-1" {
		t.Errorf("expected gateway ID gw-1, got %s", gw.ID)
	}
}

func TestGetGateway_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	gw, err := c.GetGateway(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw != nil {
		t.Errorf("expected nil gateway for 404, got %v", gw)
	}
}

func TestUpdateGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/gateways/gw-1" {
			t.Errorf("expected path /gateways/gw-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req GatewayUpdate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Name != "updated-gw" {
			t.Errorf("expected gateway name updated-gw, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Gateway{
			ID:   "gw-1",
			Name: req.Name,
			URL:  req.URL,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	gw, err := c.UpdateGateway(context.Background(), "gw-1", GatewayUpdate{
		Name: "updated-gw",
		URL:  "https://updated.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.Name != "updated-gw" {
		t.Errorf("expected gateway name updated-gw, got %s", gw.Name)
	}
}

func TestDeleteGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/gateways/gw-1" {
			t.Errorf("expected path /gateways/gw-1, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteGateway(context.Background(), "gw-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Tool Tests ---

func TestCreateTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/tools" {
			t.Errorf("expected path /tools, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req CreateToolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Tool.Name != "test-tool" {
			t.Errorf("expected tool name test-tool, got %s", req.Tool.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Tool{
			ID:          "tool-1",
			Name:        req.Tool.Name,
			Description: req.Tool.Description,
			Visibility:  req.Visibility,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	tool, err := c.CreateTool(context.Background(), CreateToolRequest{
		Tool:       ToolCreate{Name: "test-tool", Description: "Test tool"},
		Visibility: "private",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tool.ID != "tool-1" {
		t.Errorf("expected tool ID tool-1, got %s", tool.ID)
	}
	if tool.Name != "test-tool" {
		t.Errorf("expected tool name test-tool, got %s", tool.Name)
	}
}

func TestGetTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tools/tool-1" {
			t.Errorf("expected path /tools/tool-1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Tool{ID: "tool-1", Name: "test-tool"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	tool, err := c.GetTool(context.Background(), "tool-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tool == nil {
		t.Fatal("expected tool, got nil")
	}
	if tool.ID != "tool-1" {
		t.Errorf("expected tool ID tool-1, got %s", tool.ID)
	}
}

func TestGetTool_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	tool, err := c.GetTool(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tool != nil {
		t.Errorf("expected nil tool for 404, got %v", tool)
	}
}

func TestDeleteTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/tools/tool-1" {
			t.Errorf("expected path /tools/tool-1, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteTool(context.Background(), "tool-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Resource Tests ---

func TestCreateResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/resources" {
			t.Errorf("expected path /resources, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req CreateResourceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Resource.Name != "test-res" {
			t.Errorf("expected resource name test-res, got %s", req.Resource.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Resource{
			ID:         "res-1",
			URI:        req.Resource.URI,
			Name:       req.Resource.Name,
			Visibility: req.Visibility,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	res, err := c.CreateResource(context.Background(), CreateResourceRequest{
		Resource:   ResourceCreate{URI: "file:///test", Name: "test-res"},
		Visibility: "private",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != "res-1" {
		t.Errorf("expected resource ID res-1, got %s", res.ID)
	}
	if res.Name != "test-res" {
		t.Errorf("expected resource name test-res, got %s", res.Name)
	}
}

func TestGetResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/resources/res-1/info" {
			t.Errorf("expected path /resources/res-1/info, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Resource{ID: "res-1", Name: "test-res", URI: "file:///test"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	res, err := c.GetResource(context.Background(), "res-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected resource, got nil")
	}
	if res.ID != "res-1" {
		t.Errorf("expected resource ID res-1, got %s", res.ID)
	}
}

func TestGetResource_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	res, err := c.GetResource(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil resource for 404, got %v", res)
	}
}

func TestDeleteResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/resources/res-1" {
			t.Errorf("expected path /resources/res-1, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteResource(context.Background(), "res-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Prompt Tests ---

func TestCreatePrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/prompts" {
			t.Errorf("expected path /prompts, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req CreatePromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Prompt.Name != "test-prompt" {
			t.Errorf("expected prompt name test-prompt, got %s", req.Prompt.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Prompt{
			ID:          "prompt-1",
			Name:        req.Prompt.Name,
			Description: req.Prompt.Description,
			Visibility:  req.Visibility,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	prompt, err := c.CreatePrompt(context.Background(), CreatePromptRequest{
		Prompt:     PromptCreate{Name: "test-prompt", Description: "Test"},
		Visibility: "public",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.ID != "prompt-1" {
		t.Errorf("expected prompt ID prompt-1, got %s", prompt.ID)
	}
	if prompt.Name != "test-prompt" {
		t.Errorf("expected prompt name test-prompt, got %s", prompt.Name)
	}
}

func TestGetPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prompts/prompt-1" {
			t.Errorf("expected path /prompts/prompt-1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Prompt{ID: "prompt-1", Name: "test-prompt"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	prompt, err := c.GetPrompt(context.Background(), "prompt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt == nil {
		t.Fatal("expected prompt, got nil")
	}
	if prompt.ID != "prompt-1" {
		t.Errorf("expected prompt ID prompt-1, got %s", prompt.ID)
	}
}

func TestGetPrompt_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	prompt, err := c.GetPrompt(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt != nil {
		t.Errorf("expected nil prompt for 404, got %v", prompt)
	}
}

func TestDeletePrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/prompts/prompt-1" {
			t.Errorf("expected path /prompts/prompt-1, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeletePrompt(context.Background(), "prompt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Root Tests ---

func TestCreateRoot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/roots" {
			t.Errorf("expected path /roots, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req Root
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.URI != "file:///workspace" {
			t.Errorf("expected root URI file:///workspace, got %s", req.URI)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Root{
			URI:  req.URI,
			Name: req.Name,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	root, err := c.CreateRoot(context.Background(), Root{
		URI:  "file:///workspace",
		Name: "test-root",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.URI != "file:///workspace" {
		t.Errorf("expected root URI file:///workspace, got %s", root.URI)
	}
	if root.Name != "test-root" {
		t.Errorf("expected root name test-root, got %s", root.Name)
	}
}

func TestListRoots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/roots" {
			t.Errorf("expected path /roots, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected auth header, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Root{
			{URI: "file:///workspace", Name: "test-root"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	roots, err := c.ListRoots(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].URI != "file:///workspace" {
		t.Errorf("expected root URI file:///workspace, got %s", roots[0].URI)
	}
}

func TestDeleteRoot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/roots/file:///workspace" {
			t.Errorf("expected path /roots/file:///workspace, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteRoot(context.Background(), "file:///workspace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
