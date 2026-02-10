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

func TestAccServerDataSource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/servers/srv-123" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client.Server{
				ID:          "srv-123",
				Name:        "test-server",
				Description: "A test server",
				Tags:        []string{"demo"},
				Visibility:  "private",
				IsActive:    true,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"contextforge": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccServerDataSourceConfig(mockServer.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.contextforge_server.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-server"),
					),
					statecheck.ExpectKnownValue(
						"data.contextforge_server.test",
						tfjsonpath.New("visibility"),
						knownvalue.StringExact("private"),
					),
				},
			},
		},
	})
}

func testAccServerDataSourceConfig(endpoint string) string {
	return `
provider "contextforge" {
  endpoint     = "` + endpoint + `"
  bearer_token = "test"
}

data "contextforge_server" "test" {
  id = "srv-123"
}
`
}
