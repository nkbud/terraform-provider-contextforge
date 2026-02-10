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

var _ datasource.DataSource = &ToolsDataSource{}

func NewToolsDataSource() datasource.DataSource {
	return &ToolsDataSource{}
}

// ToolsDataSource lists tools from the MCP Gateway.
type ToolsDataSource struct {
	client *client.Client
}

// ToolsDataSourceModel describes the data source data model.
type ToolsDataSourceModel struct {
	IncludeInactive types.Bool      `tfsdk:"include_inactive"`
	Tools           []ToolItemModel `tfsdk:"tools"`
	ID              types.String    `tfsdk:"id"`
}

// ToolItemModel describes a single tool in the list.
type ToolItemModel struct {
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

func (d *ToolsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tools"
}

func (d *ToolsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists tools from the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"include_inactive": schema.BoolAttribute{
				MarkdownDescription: "Whether to include inactive tools in the list. Defaults to `false`.",
				Optional:            true,
			},
			"tools": schema.ListNestedAttribute{
				MarkdownDescription: "List of tools.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Tool identifier.",
							Computed:            true,
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *ToolsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ToolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ToolsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeInactive := false
	if !data.IncludeInactive.IsNull() && !data.IncludeInactive.IsUnknown() {
		includeInactive = data.IncludeInactive.ValueBool()
	}

	tools, err := d.client.ListTools(ctx, includeInactive)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list tools, got error: %s", err))
		return
	}

	data.Tools = make([]ToolItemModel, len(tools))
	for i, t := range tools {
		item := ToolItemModel{
			ID:          types.StringValue(t.ID),
			Name:        types.StringValue(t.Name),
			Description: types.StringValue(t.Description),
			IsActive:    types.BoolValue(t.IsActive),
			GatewayID:   types.StringValue(t.GatewayID),
			Visibility:  types.StringValue(t.Visibility),
			CreatedAt:   types.StringValue(t.CreatedAt),
			UpdatedAt:   types.StringValue(t.UpdatedAt),
		}

		if t.InputSchema != nil {
			schemaJSON, err := json.Marshal(t.InputSchema)
			if err != nil {
				resp.Diagnostics.AddError("InputSchema Serialization Error", fmt.Sprintf("Unable to serialize input schema: %s", err))
				return
			}
			item.InputSchema = types.StringValue(string(schemaJSON))
		} else {
			item.InputSchema = types.StringNull()
		}

		if t.Tags != nil {
			tags, diags := types.ListValueFrom(ctx, types.StringType, t.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			item.Tags = tags
		} else {
			item.Tags = types.ListNull(types.StringType)
		}

		data.Tools[i] = item
	}

	data.ID = types.StringValue("tools")

	tflog.Trace(ctx, "read tools data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
