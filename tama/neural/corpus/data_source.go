// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package corpus

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
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	Main           types.Bool   `tfsdk:"main"`
	Template       types.String `tfsdk:"template"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_class_corpus"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Neural Class Corpus",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Corpus identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the corpus",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the corpus",
				Computed:            true,
			},
			"main": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the main corpus",
				Computed:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Template for the corpus",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the corpus",
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

	// Get corpus from API
	tflog.Debug(ctx, "Reading corpus", map[string]any{
		"id": data.Id.ValueString(),
	})

	corpusResponse, err := d.client.Neural.GetCorpus(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read corpus, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(corpusResponse.ID)
	data.Name = types.StringValue(corpusResponse.Name)
	data.Slug = types.StringValue(corpusResponse.Slug)
	data.Main = types.BoolValue(corpusResponse.Main)
	data.Template = types.StringValue(corpusResponse.Template)
	data.ProvisionState = types.StringValue(corpusResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a corpus data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
