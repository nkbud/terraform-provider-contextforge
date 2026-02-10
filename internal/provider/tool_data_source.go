// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

var _ datasource.DataSource = &ToolDataSource{}

func NewToolDataSource() datasource.DataSource {
	return &ToolDataSource{}
}

// ToolDataSource reads a single tool from the MCP Gateway.
type ToolDataSource struct {
	client *client.Client
}

// ToolDataSourceModel describes the data source data model.
type ToolDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	InputSchema types.String `tfsdk:"input_schema"`
	Tags        types.List   `tfsdk:"tags"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	GatewayID   types.String `tfsdk:"gateway_id"`
	Visibility  types.String `tfsdk:"visibility"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *ToolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

func (d *ToolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single tool from the ContextForge MCP Gateway by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Tool identifier.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Tool name.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Tool description.",
				Computed:            true,
			},
			"input_schema": schema.StringAttribute{
				MarkdownDescription: "Input schema as a JSON string.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the tool.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the tool is active.",
				Computed:            true,
			},
			"gateway_id": schema.StringAttribute{
				MarkdownDescription: "Gateway ID the tool belongs to.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the tool.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the tool was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the tool was last updated.",
				Computed:            true,
			},
		},
	}
}

func (d *ToolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ToolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ToolDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tool, err := d.client.GetTool(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool, got error: %s", err))
		return
	}
	if tool == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Tool with ID %s not found", data.ID.ValueString()))
		return
	}

	data.ID = types.StringValue(tool.ID)
	data.Name = types.StringValue(tool.Name)
	data.Description = types.StringValue(tool.Description)
	data.IsActive = types.BoolValue(tool.IsActive)
	data.GatewayID = types.StringValue(tool.GatewayID)
	data.Visibility = types.StringValue(tool.Visibility)
	data.CreatedAt = types.StringValue(tool.CreatedAt)
	data.UpdatedAt = types.StringValue(tool.UpdatedAt)

	if tool.InputSchema != nil {
		schemaJSON, err := json.Marshal(tool.InputSchema)
		if err != nil {
			resp.Diagnostics.AddError("InputSchema Serialization Error", fmt.Sprintf("Unable to serialize input schema: %s", err))
			return
		}
		data.InputSchema = types.StringValue(string(schemaJSON))
	} else {
		data.InputSchema = types.StringNull()
	}

	if tool.Tags != nil {
		tags, diags := types.ListValueFrom(ctx, types.StringType, tool.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tags
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "read tool data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
