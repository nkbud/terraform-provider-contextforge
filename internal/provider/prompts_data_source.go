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

var _ datasource.DataSource = &PromptsDataSource{}

func NewPromptsDataSource() datasource.DataSource {
	return &PromptsDataSource{}
}

// PromptsDataSource lists prompts from the MCP Gateway.
type PromptsDataSource struct {
	client *client.Client
}

// PromptsDataSourceModel describes the data source data model.
type PromptsDataSourceModel struct {
	IncludeInactive types.Bool        `tfsdk:"include_inactive"`
	Prompts         []PromptItemModel `tfsdk:"prompts"`
	ID              types.String      `tfsdk:"id"`
}

// PromptItemModel describes a single prompt in the list.
type PromptItemModel struct {
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

func (d *PromptsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompts"
}

func (d *PromptsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists prompts from the ContextForge MCP Gateway.",
		Attributes: map[string]schema.Attribute{
			"include_inactive": schema.BoolAttribute{
				MarkdownDescription: "Whether to include inactive prompts in the list. Defaults to `false`.",
				Optional:            true,
			},
			"prompts": schema.ListNestedAttribute{
				MarkdownDescription: "List of prompts.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Prompt identifier.",
							Computed:            true,
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier.",
				Computed:            true,
			},
		},
	}
}

func (d *PromptsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PromptsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PromptsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeInactive := false
	if !data.IncludeInactive.IsNull() && !data.IncludeInactive.IsUnknown() {
		includeInactive = data.IncludeInactive.ValueBool()
	}

	prompts, err := d.client.ListPrompts(ctx, includeInactive)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list prompts, got error: %s", err))
		return
	}

	data.Prompts = make([]PromptItemModel, len(prompts))
	for i, p := range prompts {
		item := PromptItemModel{
			ID:          types.StringValue(p.ID),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			IsActive:    types.BoolValue(p.IsActive),
			Visibility:  types.StringValue(p.Visibility),
			CreatedAt:   types.StringValue(p.CreatedAt),
			UpdatedAt:   types.StringValue(p.UpdatedAt),
		}

		if p.Arguments != nil {
			argsJSON, err := json.Marshal(p.Arguments)
			if err != nil {
				resp.Diagnostics.AddError("Arguments Serialization Error", fmt.Sprintf("Unable to serialize arguments: %s", err))
				return
			}
			item.Arguments = types.StringValue(string(argsJSON))
		} else {
			item.Arguments = types.StringNull()
		}

		if p.Tags != nil {
			tags, diags := types.ListValueFrom(ctx, types.StringType, p.Tags)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			item.Tags = tags
		} else {
			item.Tags = types.ListNull(types.StringType)
		}

		data.Prompts[i] = item
	}

	data.ID = types.StringValue("prompts")

	tflog.Trace(ctx, "read prompts data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
