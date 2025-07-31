// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package node

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
	ClassId        types.String `tfsdk:"class_id"`
	ChainId        types.String `tfsdk:"chain_id"`
	Type           types.String `tfsdk:"type"`
	On             types.String `tfsdk:"on"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Neural Node",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Node identifier",
				Required:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this node belongs to",
				Computed:            true,
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this node uses",
				Computed:            true,
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain this node belongs to",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of node",
				Computed:            true,
			},
			"on": schema.StringAttribute{
				MarkdownDescription: "Event that triggers this node",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the node",
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

	// Get node from API
	tflog.Debug(ctx, "Reading node", map[string]any{
		"id": data.Id.ValueString(),
	})

	nodeResponse, err := d.client.Neural.GetNode(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(nodeResponse.ID)
	data.SpaceId = types.StringValue(nodeResponse.SpaceID)
	data.ClassId = types.StringValue(nodeResponse.ClassID)
	data.ChainId = types.StringValue(nodeResponse.ChainID)
	data.Type = types.StringValue(nodeResponse.Type)
	data.On = types.StringValue(nodeResponse.On)
	data.ProvisionState = types.StringValue(nodeResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a node data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
