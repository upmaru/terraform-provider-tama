// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source_identity

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

// DataSourceValidationModel describes the validation nested object for data source.
type DataSourceValidationModel struct {
	Path   types.String `tfsdk:"path"`
	Method types.String `tfsdk:"method"`
	Codes  types.List   `tfsdk:"codes"`
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	Id              types.String               `tfsdk:"id"`
	SpecificationId types.String               `tfsdk:"specification_id"`
	Identifier      types.String               `tfsdk:"identifier"`
	Validation      *DataSourceValidationModel `tfsdk:"validation"`
	ProvisionState  types.String               `tfsdk:"provision_state"`
	CurrentState    types.String               `tfsdk:"current_state"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_identity"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Sensory Source Identity",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identity identifier",
				Required:            true,
			},
			"specification_id": schema.StringAttribute{
				MarkdownDescription: "ID of the specification this identity belongs to",
				Computed:            true,
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Identifier for the identity",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current provision state of the identity",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the identity",
				Computed:            true,
			},
		},

		Blocks: map[string]schema.Block{
			"validation": schema.SingleNestedBlock{
				MarkdownDescription: "Validation configuration for the identity",
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						MarkdownDescription: "Validation endpoint path",
						Computed:            true,
					},
					"method": schema.StringAttribute{
						MarkdownDescription: "HTTP method for validation",
						Computed:            true,
					},
					"codes": schema.ListAttribute{
						MarkdownDescription: "List of acceptable HTTP status codes",
						Computed:            true,
						ElementType:         types.Int64Type,
					},
				},
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

	// Get identity from API
	tflog.Debug(ctx, "Reading source identity", map[string]any{
		"id": data.Id.ValueString(),
	})

	identityResponse, err := d.client.Sensory.GetIdentity(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source identity, got error: %s", err))
		return
	}

	// Convert response validation codes to types.List
	responseCodes := make([]int64, len(identityResponse.Validation.Codes))
	for i, code := range identityResponse.Validation.Codes {
		responseCodes[i] = int64(code)
	}
	codesList, diags := types.ListValueFrom(ctx, types.Int64Type, responseCodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(identityResponse.ID)
	data.SpecificationId = types.StringValue(identityResponse.SpecificationID)
	data.Identifier = types.StringValue(identityResponse.Identifier)
	data.ProvisionState = types.StringValue(identityResponse.ProvisionState)
	data.CurrentState = types.StringValue(identityResponse.CurrentState)
	data.Validation = &DataSourceValidationModel{
		Path:   types.StringValue(identityResponse.Validation.Path),
		Method: types.StringValue(identityResponse.Validation.Method),
		Codes:  codesList,
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a source identity data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
