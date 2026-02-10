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

func TestAccMCPResourceResource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/resources" && r.Method == http.MethodPost:
			var req client.CreateResourceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(client.Resource{
				ID:          "res-created",
				URI:         req.Resource.URI,
				Name:        req.Resource.Name,
				Description: req.Resource.Description,
				MimeType:    req.Resource.MimeType,
				Tags:        []string{},
				IsActive:    true,
				Visibility:  req.Visibility,
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case r.URL.Path == "/resources/res-created/info" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(client.Resource{
				ID:         "res-created",
				URI:        "file:///test/data.json",
				Name:       "test-res",
				MimeType:   "application/json",
				Tags:       []string{},
				IsActive:   true,
				Visibility: "private",
				CreatedAt:  "2025-01-01T00:00:00Z",
				UpdatedAt:  "2025-01-01T00:00:00Z",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case r.URL.Path == "/resources/res-created" && r.Method == http.MethodDelete:
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
				Config: testAccMCPResourceResourceConfig(mockServer.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contextforge_mcp_resource.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("res-created"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_mcp_resource.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-res"),
					),
				},
			},
		},
	})
}

func testAccMCPResourceResourceConfig(endpoint string) string {
	return `
provider "contextforge" {
  endpoint     = "` + endpoint + `"
  bearer_token = "test"
}

resource "contextforge_mcp_resource" "test" {
  uri         = "file:///test/data.json"
  name        = "test-res"
  mime_type   = "application/json"
  visibility  = "private"
}
`
}
