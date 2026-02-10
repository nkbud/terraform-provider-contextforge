// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

// Ensure ContextForgeProvider satisfies various provider interfaces.
var _ provider.Provider = &ContextForgeProvider{}
var _ provider.ProviderWithFunctions = &ContextForgeProvider{}
var _ provider.ProviderWithEphemeralResources = &ContextForgeProvider{}
var _ provider.ProviderWithActions = &ContextForgeProvider{}

// ContextForgeProvider defines the provider implementation.
type ContextForgeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ContextForgeProviderModel describes the provider data model.
type ContextForgeProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	BearerToken types.String `tfsdk:"bearer_token"`
}

func (p *ContextForgeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "contextforge"
	resp.Version = p.version
}

func (p *ContextForgeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The ContextForge provider manages resources on a ContextForge MCP Gateway instance.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "ContextForge MCP Gateway endpoint URL. Can also be set with the `CONTEXTFORGE_ENDPOINT` environment variable. Defaults to `http://localhost:4444`.",
				Optional:            true,
			},
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: "JWT bearer token for authenticating with the MCP Gateway API. Can also be set with the `MCPGATEWAY_BEARER_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ContextForgeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ContextForgeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := "http://localhost:4444"
	if !data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		endpoint = data.Endpoint.ValueString()
	} else if v := os.Getenv("CONTEXTFORGE_ENDPOINT"); v != "" {
		endpoint = v
	}

	bearerToken := ""
	if !data.BearerToken.IsNull() && !data.BearerToken.IsUnknown() {
		bearerToken = data.BearerToken.ValueString()
	} else if v := os.Getenv("MCPGATEWAY_BEARER_TOKEN"); v != "" {
		bearerToken = v
	}

	apiClient := client.NewClient(endpoint, bearerToken)
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *ContextForgeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
		NewGatewayResource,
		NewServerResource,
		NewToolResource,
		NewMCPResourceResource,
		NewPromptResource,
		NewRootResource,
	}
}

func (p *ContextForgeProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewExampleEphemeralResource,
	}
}

func (p *ContextForgeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
		NewHealthDataSource,
		NewServerDataSource,
		NewServersDataSource,
		NewGatewayDataSource,
		NewGatewaysDataSource,
		NewToolDataSource,
		NewToolsDataSource,
		NewMCPResourceDataSource,
		NewMCPResourcesDataSource,
		NewPromptDataSource,
		NewPromptsDataSource,
		NewRootsDataSource,
	}
}

func (p *ContextForgeProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func (p *ContextForgeProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{
		NewExampleAction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ContextForgeProvider{
			version: version,
		}
	}
}
