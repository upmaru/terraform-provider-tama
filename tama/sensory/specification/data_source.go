// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package specification

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
	Id             types.String `tfsdk:"id"`
	SpaceId        types.String `tfsdk:"space_id"`
	Schema         types.String `tfsdk:"schema"`
	Version        types.String `tfsdk:"version"`
	Endpoint       types.String `tfsdk:"endpoint"`
	CurrentState   types.String `tfsdk:"current_state"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_specification"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Sensory Specification",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Specification identifier",
				Required:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this specification belongs to",
				Computed:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "OpenAPI 3.0 schema definition for the specification",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the specification",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint URL for the specification",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the specification",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Provision state of the specification",
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

	// Get specification from API
	tflog.Debug(ctx, "Reading specification", map[string]any{
		"id": data.Id.ValueString(),
	})

	specResponse, err := d.client.Sensory.GetSpecification(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read specification, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(specResponse.ID)
	data.SpaceId = types.StringValue(specResponse.SpaceID)
	data.Version = types.StringValue(specResponse.Version)
	data.Endpoint = types.StringValue(specResponse.Endpoint)
	data.CurrentState = types.StringValue(specResponse.CurrentState)
	data.ProvisionState = types.StringValue(specResponse.ProvisionState)

	// Handle schema from response
	if len(specResponse.Schema) > 0 {
		schemaJSON, err := json.Marshal(specResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Serialization Error", fmt.Sprintf("Unable to serialize schema: %s", err))
			return
		}
		data.Schema = types.StringValue(string(schemaJSON))
	} else {
		data.Schema = types.StringValue("")
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a specification data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
