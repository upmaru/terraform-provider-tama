// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source

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
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Sensory Source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Source identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the source",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the source",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint URL for the source",
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

	// Get source from API
	tflog.Debug(ctx, "Reading source", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	sourceResponse, err := d.client.Sensory.GetSource(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(sourceResponse.ID)
	data.Name = types.StringValue(sourceResponse.Name)
	// Note: Type and Endpoint are not available in API response

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a source data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
