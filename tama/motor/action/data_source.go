// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

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
	ID              types.String `tfsdk:"id"`
	Identifier      types.String `tfsdk:"identifier"`
	Path            types.String `tfsdk:"path"`
	Method          types.String `tfsdk:"method"`
	SpecificationID types.String `tfsdk:"specification_id"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Motor Action",

		Attributes: map[string]schema.Attribute{
			"specification_id": schema.StringAttribute{
				MarkdownDescription: "ID of the specification this action belongs to",
				Required:            true,
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Human-readable identifier for the action",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the action",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "API endpoint path to execute",
				Computed:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "HTTP method to use for execution (GET, POST, PUT, DELETE, etc.)",
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

	// Get action from API
	tflog.Debug(ctx, "Reading action", map[string]any{
		"specification_id": data.SpecificationID.ValueString(),
		"identifier":       data.Identifier.ValueString(),
	})

	actionResponse, err := d.client.Motor.GetAction(data.SpecificationID.ValueString(), data.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read action, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.ID = types.StringValue(actionResponse.ID)
	data.Identifier = types.StringValue(actionResponse.Identifier)
	data.Path = types.StringValue(actionResponse.Path)
	data.Method = types.StringValue(actionResponse.Method)
	data.SpecificationID = types.StringValue(actionResponse.SpecificationID)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read an action data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
