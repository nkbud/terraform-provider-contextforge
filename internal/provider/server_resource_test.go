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

func TestAccServerResource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/servers" && r.Method == http.MethodPost:
			var req client.CreateServerRequest
			json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(client.Server{
				ID:          "srv-created",
				Name:        req.Server.Name,
				Description: req.Server.Description,
				Tags:        req.Server.Tags,
				Visibility:  req.Visibility,
				IsActive:    true,
			})
		case r.URL.Path == "/servers/srv-created" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client.Server{
				ID:          "srv-created",
				Name:        "my-server",
				Description: "A managed server",
				Tags:        []string{"managed"},
				Visibility:  "private",
				IsActive:    true,
			})
		case r.URL.Path == "/servers/srv-created" && r.Method == http.MethodDelete:
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
				Config: testAccServerResourceConfig(mockServer.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contextforge_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("srv-created"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_server.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("my-server"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_server.test",
						tfjsonpath.New("visibility"),
						knownvalue.StringExact("private"),
					),
				},
			},
		},
	})
}

func testAccServerResourceConfig(endpoint string) string {
	return `
provider "contextforge" {
  endpoint     = "` + endpoint + `"
  bearer_token = "test"
}

resource "contextforge_server" "test" {
  name        = "my-server"
  description = "A managed server"
  tags        = ["managed"]
  visibility  = "private"
}
`
}
