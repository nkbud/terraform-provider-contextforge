// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nkbud/terraform-provider-contextforge/internal/client"
)

var _ resource.Resource = &RootResource{}
var _ resource.ResourceWithImportState = &RootResource{}

func NewRootResource() resource.Resource {
	return &RootResource{}
}

// RootResource manages a root on the MCP Gateway.
type RootResource struct {
	client *client.Client
}

// RootResourceModel describes the resource data model.
type RootResourceModel struct {
	URI  types.String `tfsdk:"uri"`
	Name types.String `tfsdk:"name"`
}

func (r *RootResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_root"
}

func (r *RootResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a root on the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				MarkdownDescription: "URI of the root. Serves as the unique identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the root.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RootResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RootResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RootResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.Root{
		URI:  data.URI.ValueString(),
		Name: data.Name.ValueString(),
	}

	root, err := r.client.CreateRoot(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create root, got error: %s", err))
		return
	}

	data.URI = types.StringValue(root.URI)
	if root.Name != "" {
		data.Name = types.StringValue(root.Name)
	} else if data.Name.IsNull() || data.Name.IsUnknown() {
		data.Name = types.StringNull()
	}

	tflog.Trace(ctx, "created a root resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RootResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RootResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roots, err := r.client.ListRoots(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list roots, got error: %s", err))
		return
	}

	var found *client.Root
	for _, root := range roots {
		if root.URI == data.URI.ValueString() {
			found = &root
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.URI = types.StringValue(found.URI)
	if found.Name != "" {
		data.Name = types.StringValue(found.Name)
	} else {
		data.Name = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RootResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes use RequiresReplace, so Update should never be called.
	resp.Diagnostics.AddError(
		"Unexpected Update",
		"Root resources do not support in-place updates. All attribute changes require resource replacement.",
	)
}

func (r *RootResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RootResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRoot(ctx, data.URI.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete root, got error: %s", err))
		return
	}
}

func (r *RootResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uri"), req, resp)
}
