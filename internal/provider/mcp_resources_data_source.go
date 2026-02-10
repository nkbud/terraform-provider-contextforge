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

var _ datasource.DataSource = &MCPResourcesDataSource{}

func NewMCPResourcesDataSource() datasource.DataSource {
	return &MCPResourcesDataSource{}
}

// MCPResourcesDataSource lists MCP resources from the MCP Gateway.
type MCPResourcesDataSource struct {
	client *client.Client
}

// MCPResourcesDataSourceModel describes the data source data model.
type MCPResourcesDataSourceModel struct {
	IncludeInactive types.Bool             `tfsdk:"include_inactive"`
	Resources       []MCPResourceItemModel `tfsdk:"resources"`
	ID              types.String           `tfsdk:"id"`
}

// MCPResourceItemModel describes a single resource in the list.
type MCPResourceItemModel struct {
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

func (d *MCPResourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_resources"
}

func (d *MCPResourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists MCP resources from the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"include_inactive": schema.BoolAttribute{
				MarkdownDescription: "Whether to include inactive resources in the list. Defaults to `false`.",
				Optional:            true,
			},
			"resources": schema.ListNestedAttribute{
				MarkdownDescription: "List of resources.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Resource identifier.",
							Computed:            true,
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *MCPResourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MCPResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MCPResourcesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeInactive := false
	if !data.IncludeInactive.IsNull() && !data.IncludeInactive.IsUnknown() {
		includeInactive = data.IncludeInactive.ValueBool()
	}

	resources, err := d.client.ListResources(ctx, includeInactive)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list resources, got error: %s", err))
		return
	}

	data.Resources = make([]MCPResourceItemModel, len(resources))
	for i, r := range resources {
		item := MCPResourceItemModel{
			ID:          types.StringValue(r.ID),
			URI:         types.StringValue(r.URI),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
			MimeType:    types.StringValue(r.MimeType),
			IsActive:    types.BoolValue(r.IsActive),
			Visibility:  types.StringValue(r.Visibility),
			CreatedAt:   types.StringValue(r.CreatedAt),
			UpdatedAt:   types.StringValue(r.UpdatedAt),
		}

		if r.Tags != nil {
			tags, diags := types.ListValueFrom(ctx, types.StringType, r.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			item.Tags = tags
		} else {
			item.Tags = types.ListNull(types.StringType)
		}

		data.Resources[i] = item
	}

	data.ID = types.StringValue("mcp_resources")

	tflog.Trace(ctx, "read mcp_resources data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
