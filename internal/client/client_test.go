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
