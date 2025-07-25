// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chain

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
	SpaceId        types.String `tfsdk:"space_id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chain"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Perception Chain",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Chain identifier",
				Required:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this chain belongs to",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the chain",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the chain",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the chain",
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

	// Get chain from API
	tflog.Debug(ctx, "Reading chain", map[string]any{
		"id": data.Id.ValueString(),
	})

	chainResponse, err := d.client.Perception.GetChain(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read chain, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(chainResponse.ID)
	data.SpaceId = types.StringValue(chainResponse.SpaceID)
	data.Name = types.StringValue(chainResponse.Name)
	data.Slug = types.StringValue(chainResponse.Slug)
	data.ProvisionState = types.StringValue(chainResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a chain data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
