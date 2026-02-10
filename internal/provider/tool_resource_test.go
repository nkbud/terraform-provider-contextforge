// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

func TestAccToolResource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/tools" && r.Method == http.MethodPost:
			var req client.CreateToolRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(client.Tool{
				ID:          "tool-created",
				Name:        req.Tool.Name,
				Description: req.Tool.Description,
				Tags:        []string{},
				IsActive:    true,
				Visibility:  req.Visibility,
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case r.URL.Path == "/tools/tool-created" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(client.Tool{
				ID:          "tool-created",
				Name:        "test-tool",
				Description: "A test tool",
				Tags:        []string{},
				IsActive:    true,
				Visibility:  "private",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case r.URL.Path == "/tools/tool-created" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contextforge": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccToolResourceConfig(mockServer.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contextforge_tool.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("tool-created"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_tool.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-tool"),
					),
				},
			},
		},
	})
}

func testAccToolResourceConfig(endpoint string) string {
	return `
provider "contextforge" {
  endpoint     = "` + endpoint + `"
  bearer_token = "test"
}

resource "contextforge_tool" "test" {
  name        = "test-tool"
  description = "A test tool"
  visibility  = "private"
}
`
}
