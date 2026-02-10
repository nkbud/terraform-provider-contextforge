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

var _ datasource.DataSource = &GatewaysDataSource{}

func NewGatewaysDataSource() datasource.DataSource {
	return &GatewaysDataSource{}
}

// GatewaysDataSource lists gateways from the MCP Gateway.
type GatewaysDataSource struct {
	client *client.Client
}

// GatewaysDataSourceModel describes the data source data model.
type GatewaysDataSourceModel struct {
	IncludeInactive types.Bool         `tfsdk:"include_inactive"`
	Gateways        []GatewayItemModel `tfsdk:"gateways"`
	ID              types.String       `tfsdk:"id"`
}

// GatewayItemModel describes a single gateway in the list.
type GatewayItemModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	URL                 types.String `tfsdk:"url"`
	Description         types.String `tfsdk:"description"`
	Transport           types.String `tfsdk:"transport"`
	Capabilities        types.String `tfsdk:"capabilities"`
	HealthCheckURL      types.String `tfsdk:"health_check_url"`
	HealthCheckInterval types.Int64  `tfsdk:"health_check_interval"`
	HealthCheckTimeout  types.Int64  `tfsdk:"health_check_timeout"`
	HealthCheckRetries  types.Int64  `tfsdk:"health_check_retries"`
	IsActive            types.Bool   `tfsdk:"is_active"`
	Tags                types.List   `tfsdk:"tags"`
	PassthroughHeaders  types.List   `tfsdk:"passthrough_headers"`
	AuthType            types.String `tfsdk:"auth_type"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (d *GatewaysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateways"
}

func (d *GatewaysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists gateways from the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"include_inactive": schema.BoolAttribute{
				MarkdownDescription: "Whether to include inactive gateways in the list. Defaults to `false`.",
				Optional:            true,
			},
			"gateways": schema.ListNestedAttribute{
				MarkdownDescription: "List of gateways.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Gateway identifier.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Gateway name.",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "Gateway URL.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Gateway description.",
							Computed:            true,
						},
						"transport": schema.StringAttribute{
							MarkdownDescription: "Transport protocol.",
							Computed:            true,
						},
						"capabilities": schema.StringAttribute{
							MarkdownDescription: "Gateway capabilities as a JSON string.",
							Computed:            true,
						},
						"health_check_url": schema.StringAttribute{
							MarkdownDescription: "Health check URL.",
							Computed:            true,
						},
						"health_check_interval": schema.Int64Attribute{
							MarkdownDescription: "Health check interval in seconds.",
							Computed:            true,
						},
						"health_check_timeout": schema.Int64Attribute{
							MarkdownDescription: "Health check timeout in seconds.",
							Computed:            true,
						},
						"health_check_retries": schema.Int64Attribute{
							MarkdownDescription: "Health check retries.",
							Computed:            true,
						},
						"is_active": schema.BoolAttribute{
							MarkdownDescription: "Whether the gateway is active.",
							Computed:            true,
						},
						"tags": schema.ListAttribute{
							MarkdownDescription: "Tags associated with the gateway.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"passthrough_headers": schema.ListAttribute{
							MarkdownDescription: "Headers to pass through to the gateway.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"auth_type": schema.StringAttribute{
							MarkdownDescription: "Authentication type.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp when the gateway was created.",
							Computed:            true,
						},
						"updated_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp when the gateway was last updated.",
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

func (d *GatewaysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GatewaysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GatewaysDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeInactive := false
	if !data.IncludeInactive.IsNull() && !data.IncludeInactive.IsUnknown() {
		includeInactive = data.IncludeInactive.ValueBool()
	}

	gateways, err := d.client.ListGateways(ctx, includeInactive)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list gateways, got error: %s", err))
		return
	}

	data.Gateways = make([]GatewayItemModel, len(gateways))
	for i, g := range gateways {
		item := GatewayItemModel{
			ID:          types.StringValue(g.ID),
			Name:        types.StringValue(g.Name),
			URL:         types.StringValue(g.URL),
			Description: types.StringValue(g.Description),
			Transport:   types.StringValue(g.Transport),
			IsActive:    types.BoolValue(g.IsActive),
			CreatedAt:   types.StringValue(g.CreatedAt),
			UpdatedAt:   types.StringValue(g.UpdatedAt),
		}

		if g.AuthType != "" {
			item.AuthType = types.StringValue(g.AuthType)
		} else {
			item.AuthType = types.StringNull()
		}

		if g.Capabilities != nil {
			capsJSON, err := json.Marshal(g.Capabilities)
			if err != nil {
				resp.Diagnostics.AddError("Capabilities Serialization Error", fmt.Sprintf("Unable to serialize capabilities: %s", err))
				return
			}
			item.Capabilities = types.StringValue(string(capsJSON))
		} else {
			item.Capabilities = types.StringNull()
		}

		if g.HealthCheck != nil {
			item.HealthCheckURL = types.StringValue(g.HealthCheck.URL)
			item.HealthCheckInterval = types.Int64Value(int64(g.HealthCheck.Interval))
			item.HealthCheckTimeout = types.Int64Value(int64(g.HealthCheck.Timeout))
			item.HealthCheckRetries = types.Int64Value(int64(g.HealthCheck.Retries))
		} else {
			item.HealthCheckURL = types.StringNull()
			item.HealthCheckInterval = types.Int64Null()
			item.HealthCheckTimeout = types.Int64Null()
			item.HealthCheckRetries = types.Int64Null()
		}

		if g.Tags != nil {
			tags, diags := types.ListValueFrom(ctx, types.StringType, g.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			item.Tags = tags
		} else {
			item.Tags = types.ListNull(types.StringType)
		}

		if g.PassthroughHeaders != nil {
			headers, diags := types.ListValueFrom(ctx, types.StringType, g.PassthroughHeaders)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			item.PassthroughHeaders = headers
		} else {
			item.PassthroughHeaders = types.ListNull(types.StringType)
		}

		data.Gateways[i] = item
	}

	data.ID = types.StringValue("gateways")

	tflog.Trace(ctx, "read gateways data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
