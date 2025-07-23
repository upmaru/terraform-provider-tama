// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/perception"
	internalplanmodifier "github.com/upmaru/terraform-provider-tama/internal/planmodifier"
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
	Id            types.String  `tfsdk:"id"`
	ChainId       types.String  `tfsdk:"chain_id"`
	OutputClassId types.String  `tfsdk:"output_class_id"`
	Module        []ModuleModel `tfsdk:"module"`
	CurrentState  types.String  `tfsdk:"current_state"`
	Relation      types.String  `tfsdk:"relation"`
	Index         types.Int64   `tfsdk:"index"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Perception Thought",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Thought identifier",
				Required:            true,
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain this thought belongs to",
				Computed:            true,
			},
			"output_class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the output class for this thought",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the thought",
				Computed:            true,
			},
			"relation": schema.StringAttribute{
				MarkdownDescription: "Relation type for the thought",
				Computed:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Index position of the thought in the chain",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"module": schema.ListNestedBlock{
				MarkdownDescription: "Module configuration for the thought",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"reference": schema.StringAttribute{
							MarkdownDescription: "Module reference",
							Computed:            true,
						},
						"parameters": schema.StringAttribute{
							MarkdownDescription: "Module parameters as JSON string",
							Computed:            true,
						},
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

	// Get thought from API
	tflog.Debug(ctx, "Reading thought", map[string]any{
		"id": data.Id.ValueString(),
	})

	thoughtResponse, err := d.client.Perception.GetThought(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.CurrentState = types.StringValue(thoughtResponse.CurrentState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update module with response data
	err = d.updateModuleFromResponse(thoughtResponse.Module, &data)
	if err != nil {
		resp.Diagnostics.AddError("Module Error", fmt.Sprintf("Unable to update module from response: %s", err))
		return
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a thought data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModuleFromResponse updates the module block in the data source model from the API response.
func (d *DataSource) updateModuleFromResponse(responseModule perception.Module, data *DataSourceModel) error {
	moduleModel := ModuleModel{
		Reference: types.StringValue(responseModule.Reference),
	}

	// Handle parameters
	if responseModule.Parameters != nil {
		parametersJSON, err := json.Marshal(responseModule.Parameters)
		if err != nil {
			return fmt.Errorf("unable to marshal module parameters: %s", err)
		}

		// Normalize the marshaled JSON to ensure consistent formatting
		normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(parametersJSON))
		if err != nil {
			return fmt.Errorf("unable to normalize module parameters JSON: %s", err)
		}
		moduleModel.Parameters = types.StringValue(normalizedJSON)
	} else {
		moduleModel.Parameters = types.StringNull()
	}

	data.Module = []ModuleModel{moduleModel}
	return nil
}
