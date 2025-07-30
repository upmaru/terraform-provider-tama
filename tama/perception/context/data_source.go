// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package context

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
	Id             types.String `tfsdk:"id"`
	ThoughtId      types.String `tfsdk:"thought_id"`
	PromptId       types.String `tfsdk:"prompt_id"`
	Layer          types.Int64  `tfsdk:"layer"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_context"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Perception Thought Context",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Context identifier",
				Required:            true,
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this context belongs to",
				Computed:            true,
			},
			"prompt_id": schema.StringAttribute{
				MarkdownDescription: "ID of the prompt for this context",
				Computed:            true,
			},
			"layer": schema.Int64Attribute{
				MarkdownDescription: "Layer number for the context",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the context",
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

	// Get context from API
	tflog.Debug(ctx, "Reading context", map[string]any{
		"id": data.Id.ValueString(),
	})

	contextResponse, err := d.client.Perception.GetContext(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read context, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(contextResponse.ID)
	data.ThoughtId = types.StringValue(contextResponse.ThoughtID)
	data.PromptId = types.StringValue(contextResponse.PromptID)
	data.Layer = types.Int64Value(int64(contextResponse.Layer))
	data.ProvisionState = types.StringValue(contextResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a context data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
