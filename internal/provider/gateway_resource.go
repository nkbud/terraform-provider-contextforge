// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

var _ resource.Resource = &GatewayResource{}
var _ resource.ResourceWithImportState = &GatewayResource{}

func NewGatewayResource() resource.Resource {
	return &GatewayResource{}
}

// GatewayResource manages a gateway on the MCP Gateway.
type GatewayResource struct {
	client *client.Client
}

// GatewayResourceModel describes the resource data model.
type GatewayResourceModel struct {
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
	AuthValue           types.String `tfsdk:"auth_value"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (r *GatewayResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

func (r *GatewayResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a gateway on the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Gateway identifier, assigned by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the gateway.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The gateway URL.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the gateway.",
				Optional:            true,
				Computed:            true,
			},
			"transport": schema.StringAttribute{
				MarkdownDescription: "Transport protocol for the gateway (e.g. `STREAMABLEHTTP`).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("STREAMABLEHTTP", "SSE", "STDIO"),
				},
			},
			"capabilities": schema.StringAttribute{
				MarkdownDescription: "Gateway capabilities as a JSON-encoded string.",
				Optional:            true,
				Computed:            true,
			},
			"health_check_url": schema.StringAttribute{
				MarkdownDescription: "Health check URL for the gateway.",
				Optional:            true,
				Computed:            true,
			},
			"health_check_interval": schema.Int64Attribute{
				MarkdownDescription: "Health check interval in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"health_check_timeout": schema.Int64Attribute{
				MarkdownDescription: "Health check timeout in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"health_check_retries": schema.Int64Attribute{
				MarkdownDescription: "Number of health check retries.",
				Optional:            true,
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the gateway is active.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the gateway.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"passthrough_headers": schema.ListAttribute{
				MarkdownDescription: "Headers to pass through to the gateway.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type for the gateway.",
				Optional:            true,
				Computed:            true,
			},
			"auth_value": schema.StringAttribute{
				MarkdownDescription: "Authentication value for the gateway.",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *GatewayResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = apiClient
}

func (r *GatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GatewayResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var passthroughHeaders []string
	if !data.PassthroughHeaders.IsNull() && !data.PassthroughHeaders.IsUnknown() {
		resp.Diagnostics.Append(data.PassthroughHeaders.ElementsAs(ctx, &passthroughHeaders, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	isActiveCreate := true
	if !data.IsActive.IsNull() && !data.IsActive.IsUnknown() {
		isActiveCreate = data.IsActive.ValueBool()
	}

	createReq := client.GatewayCreate{
		Name:               data.Name.ValueString(),
		URL:                data.URL.ValueString(),
		Description:        data.Description.ValueString(),
		Transport:          data.Transport.ValueString(),
		IsActive:           isActiveCreate,
		Tags:               tags,
		PassthroughHeaders: passthroughHeaders,
		AuthType:           data.AuthType.ValueString(),
		AuthValue:          data.AuthValue.ValueString(),
	}

	if !data.Capabilities.IsNull() && !data.Capabilities.IsUnknown() && data.Capabilities.ValueString() != "" {
		var caps map[string]interface{}
		if err := json.Unmarshal([]byte(data.Capabilities.ValueString()), &caps); err != nil {
			resp.Diagnostics.AddError("Invalid Capabilities", fmt.Sprintf("Unable to parse capabilities JSON: %s", err))
			return
		}
		createReq.Capabilities = caps
	}

	if !data.HealthCheckURL.IsNull() && !data.HealthCheckURL.IsUnknown() {
		hc := &client.GatewayHealthCheck{
			URL: data.HealthCheckURL.ValueString(),
		}
		if !data.HealthCheckInterval.IsNull() && !data.HealthCheckInterval.IsUnknown() {
			hc.Interval = int(data.HealthCheckInterval.ValueInt64())
		}
		if !data.HealthCheckTimeout.IsNull() && !data.HealthCheckTimeout.IsUnknown() {
			hc.Timeout = int(data.HealthCheckTimeout.ValueInt64())
		}
		if !data.HealthCheckRetries.IsNull() && !data.HealthCheckRetries.IsUnknown() {
			hc.Retries = int(data.HealthCheckRetries.ValueInt64())
		}
		createReq.HealthCheck = hc
	}

	gateway, err := r.client.CreateGateway(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create gateway, got error: %s", err))
		return
	}

	r.gatewayToModel(ctx, gateway, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a gateway resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GatewayResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gateway, err := r.client.GetGateway(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read gateway, got error: %s", err))
		return
	}
	if gateway == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve auth_value from state since the API does not return it
	authValue := data.AuthValue

	r.gatewayToModel(ctx, gateway, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore auth_value — the API never echoes it back
	if !authValue.IsNull() && !authValue.IsUnknown() {
		data.AuthValue = authValue
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GatewayResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var passthroughHeaders []string
	if !data.PassthroughHeaders.IsNull() && !data.PassthroughHeaders.IsUnknown() {
		resp.Diagnostics.Append(data.PassthroughHeaders.ElementsAs(ctx, &passthroughHeaders, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	isActive := data.IsActive.ValueBool()
	updateReq := client.GatewayUpdate{
		Name:               data.Name.ValueString(),
		URL:                data.URL.ValueString(),
		Description:        data.Description.ValueString(),
		Transport:          data.Transport.ValueString(),
		IsActive:           &isActive,
		Tags:               tags,
		PassthroughHeaders: passthroughHeaders,
		AuthType:           data.AuthType.ValueString(),
		AuthValue:          data.AuthValue.ValueString(),
	}

	if !data.Capabilities.IsNull() && !data.Capabilities.IsUnknown() && data.Capabilities.ValueString() != "" {
		var caps map[string]interface{}
		if err := json.Unmarshal([]byte(data.Capabilities.ValueString()), &caps); err != nil {
			resp.Diagnostics.AddError("Invalid Capabilities", fmt.Sprintf("Unable to parse capabilities JSON: %s", err))
			return
		}
		updateReq.Capabilities = caps
	}

	if !data.HealthCheckURL.IsNull() && !data.HealthCheckURL.IsUnknown() {
		hc := &client.GatewayHealthCheck{
			URL: data.HealthCheckURL.ValueString(),
		}
		if !data.HealthCheckInterval.IsNull() && !data.HealthCheckInterval.IsUnknown() {
			hc.Interval = int(data.HealthCheckInterval.ValueInt64())
		}
		if !data.HealthCheckTimeout.IsNull() && !data.HealthCheckTimeout.IsUnknown() {
			hc.Timeout = int(data.HealthCheckTimeout.ValueInt64())
		}
		if !data.HealthCheckRetries.IsNull() && !data.HealthCheckRetries.IsUnknown() {
			hc.Retries = int(data.HealthCheckRetries.ValueInt64())
		}
		updateReq.HealthCheck = hc
	}

	gateway, err := r.client.UpdateGateway(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update gateway, got error: %s", err))
		return
	}

	// Preserve auth_value from plan since the API does not return it
	authValue := data.AuthValue

	r.gatewayToModel(ctx, gateway, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore auth_value — the API never echoes it back
	if !authValue.IsNull() && !authValue.IsUnknown() {
		data.AuthValue = authValue
	}

	tflog.Trace(ctx, "updated a gateway resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GatewayResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGateway(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete gateway, got error: %s", err))
		return
	}
}

func (r *GatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// gatewayToModel maps a client.Gateway to the Terraform resource model.
func (r *GatewayResource) gatewayToModel(ctx context.Context, gateway *client.Gateway, data *GatewayResourceModel, diagnostics *diag.Diagnostics) {
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
	if gateway.AuthValue != "" {
		data.AuthValue = types.StringValue(gateway.AuthValue)
	} else {
		data.AuthValue = types.StringNull()
	}

	if gateway.Capabilities != nil {
		capsJSON, err := json.Marshal(gateway.Capabilities)
		if err != nil {
			diagnostics.AddError("Capabilities Serialization Error", fmt.Sprintf("Unable to serialize capabilities: %s", err))
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
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, gateway.Tags)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	if gateway.PassthroughHeaders != nil {
		headersList, diags := types.ListValueFrom(ctx, types.StringType, gateway.PassthroughHeaders)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}
		data.PassthroughHeaders = headersList
	} else {
		data.PassthroughHeaders = types.ListNull(types.StringType)
	}
}
