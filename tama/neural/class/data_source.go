// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package class

import (
	"context"
	"encoding/json"
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
	Id           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	Description  types.String  `tfsdk:"description"`
	Schema       []SchemaModel `tfsdk:"schema"`
	SchemaJSON   types.String  `tfsdk:"schema_json"`
	CurrentState types.String  `tfsdk:"current_state"`
	SpaceId      types.String  `tfsdk:"space_id"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_class"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Tama Neural Class",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Class identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the class",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the class",
				Computed:            true,
			},
			"schema_json": schema.StringAttribute{
				MarkdownDescription: "JSON schema as a string",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the class",
				Computed:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this class belongs to",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"schema": schema.ListNestedBlock{
				MarkdownDescription: "JSON schema definition for the class",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							MarkdownDescription: "Title of the schema",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the schema",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the schema (e.g., 'object', 'array')",
							Computed:            true,
						},
						"properties": schema.StringAttribute{
							MarkdownDescription: "JSON string defining the properties of the schema",
							Computed:            true,
						},
						"required": schema.ListAttribute{
							MarkdownDescription: "List of required properties",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"strict": schema.BoolAttribute{
							MarkdownDescription: "Whether the schema should be strictly validated",
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

	// Get class from API
	tflog.Debug(ctx, "Reading class", map[string]any{
		"id": data.Id.ValueString(),
	})

	classResponse, err := d.client.Neural.GetClass(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read class, got error: %s", err))
		return
	}

	// Map response to data source schema
	data.Id = types.StringValue(classResponse.ID)
	data.Name = types.StringValue(classResponse.Name)
	data.Description = types.StringValue(classResponse.Description)
	data.CurrentState = types.StringValue(classResponse.CurrentState)
	data.SpaceId = types.StringValue(classResponse.SpaceID)

	// Update both schema block and schema_json with response data
	err = d.updateSchemaFromResponse(ctx, classResponse.Schema, &data)
	if err != nil {
		resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to update schema from response: %s", err))
		return
	}

	// Also populate schema_json for convenience
	schemaJSON, err := json.Marshal(classResponse.Schema)
	if err != nil {
		resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to marshal schema to JSON: %s", err))
		return
	}
	data.SchemaJSON = types.StringValue(string(schemaJSON))

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a class data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateSchemaFromResponse updates the schema block in the data source model from the API response.
func (d *DataSource) updateSchemaFromResponse(ctx context.Context, responseSchema map[string]any, data *DataSourceModel) error {
	schemaBlock := SchemaModel{}

	// Extract title
	if title, ok := responseSchema["title"].(string); ok {
		schemaBlock.Title = types.StringValue(title)
	}

	// Extract description
	if description, ok := responseSchema["description"].(string); ok {
		schemaBlock.Description = types.StringValue(description)
	}

	// Extract type
	if schemaType, ok := responseSchema["type"].(string); ok {
		schemaBlock.Type = types.StringValue(schemaType)
	}

	// Extract properties if present
	if properties, ok := responseSchema["properties"]; ok {
		propertiesJSON, err := json.Marshal(properties)
		if err != nil {
			return fmt.Errorf("unable to marshal properties: %s", err)
		}
		schemaBlock.Properties = types.StringValue(string(propertiesJSON))
	} else {
		schemaBlock.Properties = types.StringNull()
	}

	// Extract required fields if present
	if required, ok := responseSchema["required"].([]interface{}); ok {
		var requiredStrings []string
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredStrings = append(requiredStrings, reqStr)
			}
		}
		requiredList, diags := types.ListValueFrom(ctx, types.StringType, requiredStrings)
		if diags.HasError() {
			return fmt.Errorf("unable to create required list")
		}
		schemaBlock.Required = requiredList
	} else {
		schemaBlock.Required = types.ListNull(types.StringType)
	}

	// Extract strict if present
	if strict, ok := responseSchema["strict"].(bool); ok {
		schemaBlock.Strict = types.BoolValue(strict)
	} else {
		schemaBlock.Strict = types.BoolNull()
	}

	data.Schema = []SchemaModel{schemaBlock}
	return nil
}
