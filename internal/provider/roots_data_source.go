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

var _ datasource.DataSource = &RootsDataSource{}

func NewRootsDataSource() datasource.DataSource {
	return &RootsDataSource{}
}

// RootsDataSource lists roots from the MCP Gateway.
type RootsDataSource struct {
	client *client.Client
}

// RootsDataSourceModel describes the data source data model.
type RootsDataSourceModel struct {
	Roots []RootItemModel `tfsdk:"roots"`
	ID    types.String    `tfsdk:"id"`
}

// RootItemModel describes a single root in the list.
type RootItemModel struct {
	URI  types.String `tfsdk:"uri"`
	Name types.String `tfsdk:"name"`
}

func (d *RootsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roots"
}

func (d *RootsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists roots from the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"roots": schema.ListNestedAttribute{
				MarkdownDescription: "List of roots.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uri": schema.StringAttribute{
							MarkdownDescription: "Root URI.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Root name.",
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

func (d *RootsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RootsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RootsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roots, err := d.client.ListRoots(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list roots, got error: %s", err))
		return
	}

	data.Roots = make([]RootItemModel, len(roots))
	for i, r := range roots {
		data.Roots[i] = RootItemModel{
			URI:  types.StringValue(r.URI),
			Name: types.StringValue(r.Name),
		}
	}

	data.ID = types.StringValue("roots")

	tflog.Trace(ctx, "read roots data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
