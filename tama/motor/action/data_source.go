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
	"github.com/upmaru/tama-go/motor"
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
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the action",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "API endpoint path to execute. When provided with specification_id and method, will lookup action by path and method instead of identifier",
				Optional:            true,
				Computed:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "HTTP method to use for execution (GET, POST, PUT, DELETE, etc.). When provided with specification_id and path, will lookup action by path and method instead of identifier",
				Optional:            true,
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
	hasIdentifier := !data.Identifier.IsNull() && !data.Identifier.IsUnknown() && data.Identifier.ValueString() != ""
	hasPath := !data.Path.IsNull() && !data.Path.IsUnknown() && data.Path.ValueString() != ""
	hasMethod := !data.Method.IsNull() && !data.Method.IsUnknown() && data.Method.ValueString() != ""

	// Check for valid combinations
	if hasIdentifier && (hasPath || hasMethod) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Cannot specify 'identifier' together with 'path' or 'method' - use either identifier alone, or path and method together",
		)
		return
	}

	if hasPath && !hasMethod {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"When using 'path' to lookup an action, 'method' is also required",
		)
		return
	}

	if !hasPath && hasMethod {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"When using 'method' to lookup an action, 'path' is also required",
		)
		return
	}

	if !hasIdentifier && !hasPath {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Either 'identifier' or both 'path' and 'method' must be specified to lookup an action",
		)
		return
	}

	var actionResponse *motor.Action
	var err error

	if hasPath && hasMethod {
		// Get action by path and method
		tflog.Debug(ctx, "Reading action by path and method", map[string]any{
			"specification_id": data.SpecificationID.ValueString(),
			"path":             data.Path.ValueString(),
			"method":           data.Method.ValueString(),
		})
		actionResponse, err = d.client.Motor.GetActionByPathAndMethod(data.SpecificationID.ValueString(), data.Path.ValueString(), data.Method.ValueString())
	} else {
		// Get action by identifier
		tflog.Debug(ctx, "Reading action by identifier", map[string]any{
			"specification_id": data.SpecificationID.ValueString(),
			"identifier":       data.Identifier.ValueString(),
		})
		actionResponse, err = d.client.Motor.GetAction(data.SpecificationID.ValueString(), data.Identifier.ValueString())
	}

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
