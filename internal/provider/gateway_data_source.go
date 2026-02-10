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

var _ datasource.DataSource = &GatewayDataSource{}

func NewGatewayDataSource() datasource.DataSource {
	return &GatewayDataSource{}
}

// GatewayDataSource reads a single gateway from the MCP Gateway.
type GatewayDataSource struct {
	client *client.Client
}

// GatewayDataSourceModel describes the data source data model.
type GatewayDataSourceModel struct {
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

func (d *GatewayDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

func (d *GatewayDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single gateway from the ContextForge MCP Gateway by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Gateway identifier.",
				Required:            true,
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
	}
}

func (d *GatewayDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GatewayDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gateway, err := d.client.GetGateway(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read gateway, got error: %s", err))
		return
	}
	if gateway == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Gateway with ID %s not found", data.ID.ValueString()))
		return
	}

	data.ID = types.StringValue(gateway.ID)
	data.Name = types.StringValue(gateway.Name)
	data.URL = types.StringValue(gateway.URL)
	data.Description = types.StringValue(gateway.Description)
	data.Transport = types.StringValue(gateway.Transport)
	data.IsActive = types.BoolValue(gateway.IsActive)
	data.CreatedAt = types.StringValue(gateway.CreatedAt)
	data.UpdatedAt = types.StringValue(gateway.UpdatedAt)

	if gateway.AuthType != "" {
		data.AuthType = types.StringValue(gateway.AuthType)
	} else {
		data.AuthType = types.StringNull()
	}

	if gateway.Capabilities != nil {
		capsJSON, err := json.Marshal(gateway.Capabilities)
		if err != nil {
			resp.Diagnostics.AddError("Capabilities Serialization Error", fmt.Sprintf("Unable to serialize capabilities: %s", err))
			return
		}
		data.Capabilities = types.StringValue(string(capsJSON))
	} else {
		data.Capabilities = types.StringNull()
	}

	if gateway.HealthCheck != nil {
		data.HealthCheckURL = types.StringValue(gateway.HealthCheck.URL)
		data.HealthCheckInterval = types.Int64Value(int64(gateway.HealthCheck.Interval))
		data.HealthCheckTimeout = types.Int64Value(int64(gateway.HealthCheck.Timeout))
		data.HealthCheckRetries = types.Int64Value(int64(gateway.HealthCheck.Retries))
	} else {
		data.HealthCheckURL = types.StringNull()
		data.HealthCheckInterval = types.Int64Null()
		data.HealthCheckTimeout = types.Int64Null()
		data.HealthCheckRetries = types.Int64Null()
	}

	if gateway.Tags != nil {
		tags, diags := types.ListValueFrom(ctx, types.StringType, gateway.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tags
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	if gateway.PassthroughHeaders != nil {
		headers, diags := types.ListValueFrom(ctx, types.StringType, gateway.PassthroughHeaders)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.PassthroughHeaders = headers
	} else {
		data.PassthroughHeaders = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "read gateway data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
