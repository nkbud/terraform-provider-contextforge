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

func TestAccPromptResource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/prompts" && r.Method == http.MethodPost:
			var req client.CreatePromptRequest
			json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(client.Prompt{
				ID:          "prompt-created",
				Name:        req.Prompt.Name,
				Description: req.Prompt.Description,
				Tags:        []string{},
				IsActive:    true,
				Visibility:  req.Visibility,
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			})
		case r.URL.Path == "/prompts/prompt-created" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client.Prompt{
				ID:          "prompt-created",
				Name:        "test-prompt",
				Description: "A test prompt",
				Tags:        []string{},
				IsActive:    true,
				Visibility:  "public",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			})
		case r.URL.Path == "/prompts/prompt-created" && r.Method == http.MethodDelete:
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
				Config: testAccPromptResourceConfig(mockServer.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contextforge_prompt.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("prompt-created"),
					),
					statecheck.ExpectKnownValue(
						"contextforge_prompt.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-prompt"),
					),
				},
			},
		},
	})
}

func testAccPromptResourceConfig(endpoint string) string {
	return `
provider "contextforge" {
  endpoint     = "` + endpoint + `"
  bearer_token = "test"
}

resource "contextforge_prompt" "test" {
  name        = "test-prompt"
  description = "A test prompt"
  visibility  = "public"
}
`
}
