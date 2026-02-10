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

var _ resource.Resource = &PromptResource{}
var _ resource.ResourceWithImportState = &PromptResource{}

func NewPromptResource() resource.Resource {
	return &PromptResource{}
}

// PromptResource manages a prompt on the MCP Gateway.
type PromptResource struct {
	client *client.Client
}

// PromptResourceModel describes the resource data model.
type PromptResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Arguments   types.String `tfsdk:"arguments"`
	Tags        types.List   `tfsdk:"tags"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	Visibility  types.String `tfsdk:"visibility"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *PromptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (r *PromptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a prompt on the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Prompt identifier, assigned by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the prompt.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the prompt.",
				Optional:            true,
				Computed:            true,
			},
			"arguments": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded arguments array for the prompt.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the prompt.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the prompt is active.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the prompt (e.g. `public`, `private`).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "private", "team"),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the prompt was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the prompt was last updated.",
				Computed:            true,
			},
		},
	}
}

func (r *PromptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PromptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PromptResourceModel

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

	var arguments []client.PromptArgument
	if !data.Arguments.IsNull() && !data.Arguments.IsUnknown() && data.Arguments.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.Arguments.ValueString()), &arguments); err != nil {
			resp.Diagnostics.AddError("Invalid Arguments", fmt.Sprintf("Unable to parse arguments JSON: %s", err))
			return
		}
	}

	createReq := client.CreatePromptRequest{
		Prompt: client.PromptCreate{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			Arguments:   arguments,
			Tags:        tags,
		},
		Visibility: data.Visibility.ValueString(),
	}

	prompt, err := r.client.CreatePrompt(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create prompt, got error: %s", err))
		return
	}

	r.promptToModel(ctx, prompt, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a prompt resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PromptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PromptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prompt, err := r.client.GetPrompt(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prompt, got error: %s", err))
		return
	}
	if prompt == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.promptToModel(ctx, prompt, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PromptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PromptResourceModel

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

	var arguments []client.PromptArgument
	if !data.Arguments.IsNull() && !data.Arguments.IsUnknown() && data.Arguments.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.Arguments.ValueString()), &arguments); err != nil {
			resp.Diagnostics.AddError("Invalid Arguments", fmt.Sprintf("Unable to parse arguments JSON: %s", err))
			return
		}
	}

	updateReq := client.PromptUpdate{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Arguments:   arguments,
		Tags:        tags,
	}

	prompt, err := r.client.UpdatePrompt(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update prompt, got error: %s", err))
		return
	}

	r.promptToModel(ctx, prompt, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a prompt resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PromptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PromptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePrompt(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete prompt, got error: %s", err))
		return
	}
}

func (r *PromptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// promptToModel maps a client.Prompt to the Terraform resource model.
func (r *PromptResource) promptToModel(ctx context.Context, prompt *client.Prompt, data *PromptResourceModel, diagnostics *diag.Diagnostics) {
	data.ID = types.StringValue(prompt.ID)
	data.Name = types.StringValue(prompt.Name)
	data.Description = types.StringValue(prompt.Description)
	data.IsActive = types.BoolValue(prompt.IsActive)
	data.Visibility = types.StringValue(prompt.Visibility)
	data.CreatedAt = types.StringValue(prompt.CreatedAt)
	data.UpdatedAt = types.StringValue(prompt.UpdatedAt)

	if prompt.Arguments != nil {
		argumentsJSON, err := json.Marshal(prompt.Arguments)
		if err != nil {
			diagnostics.AddError("Serialization Error", fmt.Sprintf("Unable to serialize arguments to JSON: %s", err))
			return
		}
		data.Arguments = types.StringValue(string(argumentsJSON))
	} else {
		data.Arguments = types.StringNull()
	}

	if prompt.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, prompt.Tags)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}
}
