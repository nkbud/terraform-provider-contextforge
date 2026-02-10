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

func TestAccGatewayResource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/gateways" && r.Method == http.MethodPost:
			var req client.GatewayCreate
			json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(client.Gateway{
				ID:                 "gw-created",
				Name:               req.Name,
				URL:                req.URL,
				Description:        req.Description,
				Transport:          req.Transport,
				IsActive:           req.IsActive,
				Tags:               req.Tags,
				PassthroughHeaders: []string{},
				CreatedAt:          "2025-01-01T00:00:00Z",
				UpdatedAt:          "2025-01-01T00:00:00Z",
			})
		case r.URL.Path == "/gateways/gw-created" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client.Gateway{
				ID:                 "gw-created",
				Name:               "test-gw",
				URL:                "https://example.com/mcp",
				Transport:          "STREAMABLEHTTP",
				IsActive:           true,
				Tags:               []string{"test"},
				PassthroughHeaders: []string{},
				CreatedAt:          "2025-01-01T00:00:00Z",
				UpdatedAt:          "2025-01-01T00:00:00Z",
			})
		case r.URL.Path == "/gateways/gw-created" && r.Method == http.MethodDelete:
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
				Config: testAccGatewayResourceConfig(mockServer.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contextforge_gateway.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("gw-created"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_gateway.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-gw"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_gateway.test",
						tfjsonpath.New("transport"),
						knownvalue.StringExact("STREAMABLEHTTP"),
					),
				},
			},
		},
	})
}

func testAccGatewayResourceConfig(endpoint string) string {
	return `
provider "contextforge" {
  endpoint     = "` + endpoint + `"
  bearer_token = "test"
}

resource "contextforge_gateway" "test" {
  name      = "test-gw"
  url       = "https://example.com/mcp"
  transport = "STREAMABLEHTTP"
  is_active = true
  tags      = ["test"]
}
`
}
