// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package prompt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource defines the data source implementation.
type DataSource struct {
	client *tama.Client
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	SpaceId      types.String `tfsdk:"space_id"`
	Slug         types.String `tfsdk:"slug"`
	Content      types.String `tfsdk:"content"`
	Role         types.String `tfsdk:"role"`
	CurrentState types.String `tfsdk:"current_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Memory Prompt",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Prompt identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the prompt",
				Computed:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this prompt belongs to",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug for the prompt",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Content of the prompt",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role associated with the prompt (system or user)",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the prompt",
				Computed:            true,
			},
		},
	}
}

func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tama.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *tama.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get prompt from API
	tflog.Debug(ctx, "Reading prompt", map[string]any{
		"id": data.Id.ValueString(),
	})

	promptResponse, err := d.client.Memory.GetPrompt(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prompt, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(promptResponse.ID)
	data.Name = types.StringValue(promptResponse.Name)
	data.SpaceId = types.StringValue(promptResponse.SpaceID)
	data.Slug = types.StringValue(promptResponse.Slug)
	data.Content = types.StringValue(promptResponse.Content)
	data.Role = types.StringValue(promptResponse.Role)
	data.CurrentState = types.StringValue(promptResponse.CurrentState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a prompt data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
