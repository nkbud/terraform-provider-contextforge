// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

var _ datasource.DataSource = &HealthDataSource{}

func NewHealthDataSource() datasource.DataSource {
	return &HealthDataSource{}
}

// HealthDataSource reads the health status from the MCP Gateway.
type HealthDataSource struct {
	client *client.Client
}

// HealthDataSourceModel describes the data source data model.
type HealthDataSourceModel struct {
	Status types.String `tfsdk:"status"`
	ID     types.String `tfsdk:"id"`
}

func (d *HealthDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_health"
}

func (d *HealthDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the health status of the ContextForge MCP Gateway. No authentication required.",
		Attributes: map[string]schema.Attribute{
			"status": schema.StringAttribute{
				MarkdownDescription: "Health status of the MCP Gateway.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *HealthDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = apiClient
}

func (d *HealthDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HealthDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	health, err := d.client.GetHealth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read health, got error: %s", err))
		return
	}

	data.Status = types.StringValue(health.Status)
	data.ID = types.StringValue("health")

	tflog.Trace(ctx, "read health data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
