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
	"github.com/upmaru/tama-go/neural"
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
	ClassId        types.String `tfsdk:"class_id"`
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
		MarkdownDescription: "Fetches information about a Tama Neural Class Corpus. Can be fetched by ID directly or by class_id and slug.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Corpus identifier. Optional if class_id and slug are provided.",
				Optional:            true,
				Computed:            true,
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this corpus belongs to. Required if using slug to find the corpus.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the corpus",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the corpus. Required if using class_id to find the corpus.",
				Optional:            true,
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

	// Validate input parameters
	hasId := !data.Id.IsNull() && !data.Id.IsUnknown() && data.Id.ValueString() != ""
	hasClassIdAndSlug := (!data.ClassId.IsNull() && !data.ClassId.IsUnknown() && data.ClassId.ValueString() != "") &&
		(!data.Slug.IsNull() && !data.Slug.IsUnknown() && data.Slug.ValueString() != "")

	if !hasId && !hasClassIdAndSlug {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Either 'id' or both 'class_id' and 'slug' must be provided",
		)
		return
	}

	if hasId && hasClassIdAndSlug {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Cannot provide both 'id' and 'class_id'/'slug' simultaneously. Use one method or the other.",
		)
		return
	}

	var corpusResponse *neural.Corpus
	var err error

	if hasId {
		// Get corpus by ID
		tflog.Debug(ctx, "Reading corpus by ID", map[string]any{
			"id": data.Id.ValueString(),
		})

		corpusResponse, err = d.client.Neural.GetCorpus(data.Id.ValueString())
	} else {
		// Get corpus by class ID and slug
		tflog.Debug(ctx, "Reading corpus by class and slug", map[string]any{
			"class_id": data.ClassId.ValueString(),
			"slug":     data.Slug.ValueString(),
		})

		corpusResponse, err = d.client.Neural.GetCorpusByClassAndSlug(
			data.ClassId.ValueString(),
			data.Slug.ValueString(),
		)
	}

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
