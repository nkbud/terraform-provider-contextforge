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

var _ datasource.DataSource = &PromptDataSource{}

func NewPromptDataSource() datasource.DataSource {
	return &PromptDataSource{}
}

// PromptDataSource reads a single prompt from the MCP Gateway.
type PromptDataSource struct {
	client *client.Client
}

// PromptDataSourceModel describes the data source data model.
type PromptDataSourceModel struct {
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

func (d *PromptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (d *PromptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single prompt from the ContextForge MCP Gateway by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Prompt identifier.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Prompt name.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Prompt description.",
				Computed:            true,
			},
			"arguments": schema.StringAttribute{
				MarkdownDescription: "Prompt arguments as a JSON string.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the prompt.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the prompt is active.",
				Computed:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Visibility of the prompt.",
				Computed:            true,
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

func (d *PromptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PromptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PromptDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prompt, err := d.client.GetPrompt(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prompt, got error: %s", err))
		return
	}
	if prompt == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Prompt with ID %s not found", data.ID.ValueString()))
		return
	}

	data.ID = types.StringValue(prompt.ID)
	data.Name = types.StringValue(prompt.Name)
	data.Description = types.StringValue(prompt.Description)
	data.IsActive = types.BoolValue(prompt.IsActive)
	data.Visibility = types.StringValue(prompt.Visibility)
	data.CreatedAt = types.StringValue(prompt.CreatedAt)
	data.UpdatedAt = types.StringValue(prompt.UpdatedAt)

	if prompt.Arguments != nil {
		argsJSON, err := json.Marshal(prompt.Arguments)
		if err != nil {
			resp.Diagnostics.AddError("Arguments Serialization Error", fmt.Sprintf("Unable to serialize arguments: %s", err))
			return
		}
		data.Arguments = types.StringValue(string(argsJSON))
	} else {
		data.Arguments = types.StringNull()
	}

	if prompt.Tags != nil {
		tags, diags := types.ListValueFrom(ctx, types.StringType, prompt.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tags
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "read prompt data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
