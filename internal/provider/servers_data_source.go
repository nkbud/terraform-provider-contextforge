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

var _ datasource.DataSource = &ServersDataSource{}

func NewServersDataSource() datasource.DataSource {
	return &ServersDataSource{}
}

// ServersDataSource lists servers from the MCP Gateway.
type ServersDataSource struct {
	client *client.Client
}

// ServersDataSourceModel describes the data source data model.
type ServersDataSourceModel struct {
	IncludeInactive types.Bool            `tfsdk:"include_inactive"`
	Servers         []ServerItemModel     `tfsdk:"servers"`
	ID              types.String          `tfsdk:"id"`
}

// ServerItemModel describes a single server in the list.
type ServerItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Visibility  types.String `tfsdk:"visibility"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *ServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists servers from the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"include_inactive": schema.BoolAttribute{
				MarkdownDescription: "Whether to include inactive servers in the list. Defaults to `false`.",
				Optional:            true,
			},
			"servers": schema.ListNestedAttribute{
				MarkdownDescription: "List of servers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Server identifier.",
							Computed:            true,
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
						"visibility": schema.StringAttribute{
							MarkdownDescription: "Visibility of the server.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Server status.",
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *ServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeInactive := false
	if !data.IncludeInactive.IsNull() && !data.IncludeInactive.IsUnknown() {
		includeInactive = data.IncludeInactive.ValueBool()
	}

	servers, err := d.client.ListServers(ctx, includeInactive)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list servers, got error: %s", err))
		return
	}

	data.Servers = make([]ServerItemModel, len(servers))
	for i, s := range servers {
		tags, diags := types.ListValueFrom(ctx, types.StringType, s.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Servers[i] = ServerItemModel{
			ID:          types.StringValue(s.ID),
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
			Tags:        tags,
			Visibility:  types.StringValue(s.Visibility),
			Status:      types.StringValue(s.Status),
			CreatedAt:   types.StringValue(s.CreatedAt),
			UpdatedAt:   types.StringValue(s.UpdatedAt),
		}
	}

	data.ID = types.StringValue("servers")

	tflog.Trace(ctx, "read servers data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
