// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
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

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

// ServerResource manages a server on the MCP Gateway.
type ServerResource struct {
	client *client.Client
}

// ServerResourceModel describes the resource data model.
type ServerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	ToolIDs     types.List   `tfsdk:"tool_ids"`
	Visibility  types.String `tfsdk:"visibility"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a server on the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier, assigned by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the server.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the server.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the server.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"tool_ids": schema.ListAttribute{
				MarkdownDescription: "List of tool IDs associated with the server.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the server (e.g. `public`, `private`).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "private", "team"),
				},
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the server is active.",
				Optional:            true,
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
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServerResourceModel

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

	createReq := client.CreateServerRequest{
		Server: client.ServerConfig{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			Tags:        tags,
		},
		Visibility: data.Visibility.ValueString(),
	}

	server, err := r.client.CreateServer(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create server, got error: %s", err))
		return
	}

	r.serverToModel(ctx, server, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a server resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server, got error: %s", err))
		return
	}
	if server == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.serverToModel(ctx, server, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServerResourceModel

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

	var toolIDs []string
	if !data.ToolIDs.IsNull() && !data.ToolIDs.IsUnknown() {
		resp.Diagnostics.Append(data.ToolIDs.ElementsAs(ctx, &toolIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateReq := client.ServerUpdate{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Tags:        tags,
		ToolIDs:     toolIDs,
	}

	server, err := r.client.UpdateServer(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server, got error: %s", err))
		return
	}

	r.serverToModel(ctx, server, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a server resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServer(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete server, got error: %s", err))
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// serverToModel maps a client.Server to the Terraform resource model.
func (r *ServerResource) serverToModel(ctx context.Context, server *client.Server, data *ServerResourceModel, diagnostics *diag.Diagnostics) {
	data.ID = types.StringValue(server.ID)
	data.Name = types.StringValue(server.Name)
	data.Description = types.StringValue(server.Description)
	data.Visibility = types.StringValue(server.Visibility)
	data.IsActive = types.BoolValue(server.IsActive)
	data.CreatedAt = types.StringValue(server.CreatedAt)
	data.UpdatedAt = types.StringValue(server.UpdatedAt)

	if server.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, server.Tags)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	if server.ToolIDs != nil {
		toolIDsList, diags := types.ListValueFrom(ctx, types.StringType, server.ToolIDs)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}
		data.ToolIDs = toolIDsList
	} else {
		data.ToolIDs = types.ListNull(types.StringType)
	}
}
