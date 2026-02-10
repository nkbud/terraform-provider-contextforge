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

var _ datasource.DataSource = &ServerDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

// ServerDataSource reads a single server from the MCP Gateway.
type ServerDataSource struct {
	client *client.Client
}

// ServerDataSourceModel describes the data source data model.
type ServerDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	ToolIDs     types.List   `tfsdk:"tool_ids"`
	Visibility  types.String `tfsdk:"visibility"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single server from the ContextForge MCP Gateway by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Server description.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the server.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"tool_ids": schema.ListAttribute{
				MarkdownDescription: "List of tool IDs associated with the server.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the server (e.g. `public`, `private`).",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the server is active.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the server was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the server was last updated.",
				Computed:            true,
			},
		},
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := d.client.GetServer(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server, got error: %s", err))
		return
	}
	if server == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Server with ID %s not found", data.ID.ValueString()))
		return
	}

	data.ID = types.StringValue(server.ID)
	data.Name = types.StringValue(server.Name)
	data.Description = types.StringValue(server.Description)
	data.Visibility = types.StringValue(server.Visibility)
	data.IsActive = types.BoolValue(server.IsActive)
	data.CreatedAt = types.StringValue(server.CreatedAt)
	data.UpdatedAt = types.StringValue(server.UpdatedAt)

	tags, diags := types.ListValueFrom(ctx, types.StringType, server.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tags

	toolIDs, diags := types.ListValueFrom(ctx, types.StringType, server.ToolIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ToolIDs = toolIDs

	tflog.Trace(ctx, "read server data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
