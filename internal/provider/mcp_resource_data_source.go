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

var _ datasource.DataSource = &MCPResourceDataSource{}

func NewMCPResourceDataSource() datasource.DataSource {
	return &MCPResourceDataSource{}
}

// MCPResourceDataSource reads a single MCP resource from the MCP Gateway.
type MCPResourceDataSource struct {
	client *client.Client
}

// MCPResourceDataSourceModel describes the data source data model.
type MCPResourceDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	URI         types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	MimeType    types.String `tfsdk:"mime_type"`
	Tags        types.List   `tfsdk:"tags"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	Visibility  types.String `tfsdk:"visibility"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *MCPResourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_resource"
}

func (d *MCPResourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single MCP resource from the ContextForge MCP Gateway by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Resource URI.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Resource description.",
				Computed:            true,
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "MIME type of the resource.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the resource.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the resource is active.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the resource.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the resource was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the resource was last updated.",
				Computed:            true,
			},
		},
	}
}

func (d *MCPResourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MCPResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MCPResourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := d.client.GetResource(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read resource, got error: %s", err))
		return
	}
	if resource == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Resource with ID %s not found", data.ID.ValueString()))
		return
	}

	data.ID = types.StringValue(resource.ID)
	data.URI = types.StringValue(resource.URI)
	data.Name = types.StringValue(resource.Name)
	data.Description = types.StringValue(resource.Description)
	data.MimeType = types.StringValue(resource.MimeType)
	data.IsActive = types.BoolValue(resource.IsActive)
	data.Visibility = types.StringValue(resource.Visibility)
	data.CreatedAt = types.StringValue(resource.CreatedAt)
	data.UpdatedAt = types.StringValue(resource.UpdatedAt)

	if resource.Tags != nil {
		tags, diags := types.ListValueFrom(ctx, types.StringType, resource.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tags
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "read mcp_resource data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
