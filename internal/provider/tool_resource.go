// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

var _ resource.Resource = &ToolResource{}
var _ resource.ResourceWithImportState = &ToolResource{}

func NewToolResource() resource.Resource {
	return &ToolResource{}
}

// ToolResource manages a tool on the MCP Gateway.
type ToolResource struct {
	client *client.Client
}

// ToolResourceModel describes the resource data model.
type ToolResourceModel struct {
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

func (r *ToolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

func (r *ToolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a tool on the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Tool identifier, assigned by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the tool.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the tool.",
				Optional:            true,
				Computed:            true,
			},
			"input_schema": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded input schema for the tool.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the tool.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the tool is active.",
				Computed:            true,
			},
			"gateway_id": schema.StringAttribute{
				MarkdownDescription: "Gateway ID associated with the tool.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the tool (e.g. `public`, `private`).",
				Optional:            true,
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

func (r *ToolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ToolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ToolResourceModel

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

	var inputSchema map[string]interface{}
	if !data.InputSchema.IsNull() && !data.InputSchema.IsUnknown() && data.InputSchema.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.InputSchema.ValueString()), &inputSchema); err != nil {
			resp.Diagnostics.AddError("Invalid Input Schema", fmt.Sprintf("Unable to parse input_schema JSON: %s", err))
			return
		}
	}

	createReq := client.CreateToolRequest{
		Tool: client.ToolCreate{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			InputSchema: inputSchema,
			Tags:        tags,
		},
		Visibility: data.Visibility.ValueString(),
	}

	tool, err := r.client.CreateTool(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tool, got error: %s", err))
		return
	}

	r.toolToModel(ctx, tool, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a tool resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ToolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ToolResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tool, err := r.client.GetTool(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool, got error: %s", err))
		return
	}
	if tool == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.toolToModel(ctx, tool, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ToolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ToolResourceModel

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

	var inputSchema map[string]interface{}
	if !data.InputSchema.IsNull() && !data.InputSchema.IsUnknown() && data.InputSchema.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.InputSchema.ValueString()), &inputSchema); err != nil {
			resp.Diagnostics.AddError("Invalid Input Schema", fmt.Sprintf("Unable to parse input_schema JSON: %s", err))
			return
		}
	}

	updateReq := client.ToolUpdate{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		InputSchema: inputSchema,
		Tags:        tags,
	}

	tool, err := r.client.UpdateTool(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tool, got error: %s", err))
		return
	}

	r.toolToModel(ctx, tool, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a tool resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ToolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ToolResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTool(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tool, got error: %s", err))
		return
	}
}

func (r *ToolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// toolToModel maps a client.Tool to the Terraform resource model.
func (r *ToolResource) toolToModel(ctx context.Context, tool *client.Tool, data *ToolResourceModel, diagnostics *diag.Diagnostics) {
	data.ID = types.StringValue(tool.ID)
	data.Name = types.StringValue(tool.Name)
	data.Description = types.StringValue(tool.Description)
	data.IsActive = types.BoolValue(tool.IsActive)
	data.GatewayID = types.StringValue(tool.GatewayID)
	data.Visibility = types.StringValue(tool.Visibility)
	data.CreatedAt = types.StringValue(tool.CreatedAt)
	data.UpdatedAt = types.StringValue(tool.UpdatedAt)

	if tool.InputSchema != nil {
		inputSchemaJSON, err := json.Marshal(tool.InputSchema)
		if err != nil {
			diagnostics.AddError("Serialization Error", fmt.Sprintf("Unable to serialize input_schema to JSON: %s", err))
			return
		}
		data.InputSchema = types.StringValue(string(inputSchemaJSON))
	} else {
		data.InputSchema = types.StringNull()
	}

	tagsList, diags := types.ListValueFrom(ctx, types.StringType, tool.Tags)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
	data.Tags = tagsList
}
