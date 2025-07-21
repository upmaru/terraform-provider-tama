// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package model

import (
	"context"
	"encoding/json"
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
	Id         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Path       types.String `tfsdk:"path"`
	Parameters types.String `tfsdk:"parameters"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Sensory Model",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Model identifier",
				Required:            true,
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Model identifier (e.g., 'mistral-small-latest')",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "API path for the model (e.g., '/chat/completions')",
				Computed:            true,
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "Model parameters as JSON string",
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

	// Get model from API
	tflog.Debug(ctx, "Reading model", map[string]any{
		"id": data.Id.ValueString(),
	})

	modelResponse, err := d.client.Sensory.GetModel(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read model, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(modelResponse.ID)
	data.Identifier = types.StringValue(modelResponse.Identifier)
	// Note: Path is not available in API response

	// Handle parameters from response
	if len(modelResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(modelResponse.Parameters)
		if err != nil {
			resp.Diagnostics.AddError("Parameters Serialization Error", fmt.Sprintf("Unable to serialize parameters: %s", err))
			return
		}
		data.Parameters = types.StringValue(string(parametersJSON))
	} else {
		data.Parameters = types.StringValue("")
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a model data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
