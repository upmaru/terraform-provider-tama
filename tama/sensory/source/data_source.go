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
	"github.com/upmaru/tama-go/sensory"
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
	Id              types.String `tfsdk:"id"`
	SpecificationId types.String `tfsdk:"specification_id"`
	Slug            types.String `tfsdk:"slug"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Endpoint        types.String `tfsdk:"endpoint"`
	SpaceId         types.String `tfsdk:"space_id"`
	ProvisionState  types.String `tfsdk:"provision_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Sensory Source. Can be fetched by ID directly or by specification_id and slug.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Source identifier. Optional if specification_id and slug are provided.",
				Optional:            true,
				Computed:            true,
			},
			"specification_id": schema.StringAttribute{
				MarkdownDescription: "Specification identifier. Required if using slug to find the source.",
				Optional:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Source slug. Required if using specification_id to find the source.",
				Optional:            true,
				Computed:            true,
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
			"space_id": schema.StringAttribute{
				MarkdownDescription: "Space identifier",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Provision state of the source",
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
	hasSpecificationIdAndSlug := (!data.SpecificationId.IsNull() && !data.SpecificationId.IsUnknown() && data.SpecificationId.ValueString() != "") &&
		(!data.Slug.IsNull() && !data.Slug.IsUnknown() && data.Slug.ValueString() != "")

	if !hasId && !hasSpecificationIdAndSlug {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Either 'id' or both 'specification_id' and 'slug' must be provided",
		)
		return
	}

	if hasId && hasSpecificationIdAndSlug {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Cannot provide both 'id' and 'specification_id'/'slug' simultaneously. Use one method or the other.",
		)
		return
	}

	var sourceResponse *sensory.Source
	var err error

	if hasId {
		// Get source by ID
		tflog.Debug(ctx, "Reading source by ID", map[string]any{
			"id": data.Id.ValueString(),
		})

		sourceResponse, err = d.client.Sensory.GetSource(data.Id.ValueString())
	} else {
		// Get source by specification ID and slug
		tflog.Debug(ctx, "Reading source by specification and slug", map[string]any{
			"specification_id": data.SpecificationId.ValueString(),
			"slug":             data.Slug.ValueString(),
		})

		sourceResponse, err = d.client.Sensory.GetSourceBySpecificationAndSlug(
			data.SpecificationId.ValueString(),
			data.Slug.ValueString(),
		)
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(sourceResponse.ID)
	data.Name = types.StringValue(sourceResponse.Name)
	data.Slug = types.StringValue(sourceResponse.Slug)
	data.Type = types.StringValue(sourceResponse.Type)
	data.Endpoint = types.StringValue(sourceResponse.Endpoint)
	data.SpaceId = types.StringValue(sourceResponse.SpaceID)
	data.ProvisionState = types.StringValue(sourceResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a source data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
