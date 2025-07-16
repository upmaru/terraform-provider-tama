// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package limit

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
	Id         types.String `tfsdk:"id"`
	SourceId   types.String `tfsdk:"source_id"`
	ScaleUnit  types.String `tfsdk:"scale_unit"`
	ScaleCount types.Int64  `tfsdk:"scale_count"`
	Limit      types.Int64  `tfsdk:"limit"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_limit"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Sensory Limit",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Limit identifier",
				Required:            true,
			},
			"source_id": schema.StringAttribute{
				MarkdownDescription: "ID of the source this limit belongs to",
				Computed:            true,
			},
			"scale_unit": schema.StringAttribute{
				MarkdownDescription: "Unit for the scaling period (e.g., 'seconds', 'minutes', 'hours')",
				Computed:            true,
			},
			"scale_count": schema.Int64Attribute{
				MarkdownDescription: "Number of scale units for the limit period",
				Computed:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "The limit value for the specified period",
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

	// Get limit from API
	tflog.Debug(ctx, "Reading limit", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	limitResponse, err := d.client.Sensory.GetLimit(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read limit, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(limitResponse.ID)
	data.SourceId = types.StringValue(limitResponse.SourceID)
	data.ScaleUnit = types.StringValue(limitResponse.ScaleUnit)
	data.ScaleCount = types.Int64Value(int64(limitResponse.ScaleCount))
	data.Limit = types.Int64Value(int64(limitResponse.Count))

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a limit data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
