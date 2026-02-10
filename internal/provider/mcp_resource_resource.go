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

var _ resource.Resource = &MCPResourceResource{}
var _ resource.ResourceWithImportState = &MCPResourceResource{}

func NewMCPResourceResource() resource.Resource {
	return &MCPResourceResource{}
}

// MCPResourceResource manages an MCP resource on the MCP Gateway.
type MCPResourceResource struct {
	client *client.Client
}

// MCPResourceResourceModel describes the resource data model.
type MCPResourceResourceModel struct {
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

func (r *MCPResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_resource"
}

func (r *MCPResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an MCP resource on the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "MCP resource identifier, assigned by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "URI of the MCP resource.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the MCP resource.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the MCP resource.",
				Optional:            true,
				Computed:            true,
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "MIME type of the MCP resource.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the MCP resource.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the MCP resource is active.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the MCP resource (e.g. `public`, `private`).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "private", "team"),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the MCP resource was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the MCP resource was last updated.",
				Computed:            true,
			},
		},
	}
}

func (r *MCPResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MCPResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MCPResourceResourceModel

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

	createReq := client.CreateResourceRequest{
		Resource: client.ResourceCreate{
			URI:         data.URI.ValueString(),
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			MimeType:    data.MimeType.ValueString(),
			Tags:        tags,
		},
		Visibility: data.Visibility.ValueString(),
	}

	mcpResource, err := r.client.CreateResource(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create MCP resource, got error: %s", err))
		return
	}

	r.resourceToModel(ctx, mcpResource, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created an MCP resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MCPResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MCPResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mcpResource, err := r.client.GetResource(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read MCP resource, got error: %s", err))
		return
	}
	if mcpResource == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.resourceToModel(ctx, mcpResource, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MCPResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MCPResourceResourceModel

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

	updateReq := client.ResourceUpdate{
		URI:         data.URI.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		MimeType:    data.MimeType.ValueString(),
		Tags:        tags,
	}

	mcpResource, err := r.client.UpdateResource(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update MCP resource, got error: %s", err))
		return
	}

	r.resourceToModel(ctx, mcpResource, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated an MCP resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MCPResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MCPResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete MCP resource, got error: %s", err))
		return
	}
}

func (r *MCPResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// resourceToModel maps a client.Resource to the Terraform resource model.
func (r *MCPResourceResource) resourceToModel(ctx context.Context, mcpResource *client.Resource, data *MCPResourceResourceModel, diagnostics *diag.Diagnostics) {
	data.ID = types.StringValue(mcpResource.ID)
	data.URI = types.StringValue(mcpResource.URI)
	data.Name = types.StringValue(mcpResource.Name)
	data.Description = types.StringValue(mcpResource.Description)
	data.MimeType = types.StringValue(mcpResource.MimeType)
	data.IsActive = types.BoolValue(mcpResource.IsActive)
	data.Visibility = types.StringValue(mcpResource.Visibility)
	data.CreatedAt = types.StringValue(mcpResource.CreatedAt)
	data.UpdatedAt = types.StringValue(mcpResource.UpdatedAt)

	if mcpResource.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, mcpResource.Tags)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}
}
