// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package class

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/neural"
	internalplanmodifier "github.com/upmaru/terraform-provider-tama/internal/planmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *tama.Client
}

// SchemaModel describes the schema block data model.
type SchemaModel struct {
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Properties  types.String `tfsdk:"properties"`
	Required    types.List   `tfsdk:"required"`
	Strict      types.Bool   `tfsdk:"strict"`
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Description    types.String  `tfsdk:"description"`
	Schema         []SchemaModel `tfsdk:"schema"`
	SchemaJSON     types.String  `tfsdk:"schema_json"`
	ProvisionState types.String  `tfsdk:"provision_state"`
	SpaceId        types.String  `tfsdk:"space_id"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_class"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Class resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Class identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				MarkdownDescription: "JSON schema as a string. Mutually exclusive with schema block.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					internalplanmodifier.JSONNormalize(),
				},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the class",
				Computed:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this class belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"schema": schema.ListNestedBlock{
				MarkdownDescription: "JSON schema definition for the class. Mutually exclusive with schema_json attribute.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							MarkdownDescription: "Title of the schema",
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the schema",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the schema (e.g., 'object', 'array')",
							Required:            true,
						},
						"properties": schema.StringAttribute{
							MarkdownDescription: "JSON string defining the properties of the schema",
							Optional:            true,
							PlanModifiers: []planmodifier.String{
								internalplanmodifier.JSONNormalize(),
							},
						},
						"required": schema.ListAttribute{
							MarkdownDescription: "List of required properties",
							Optional:            true,
							ElementType:         types.StringType,
						},
						"strict": schema.BoolAttribute{
							MarkdownDescription: "Whether the schema should be strictly validated",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tama.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tama.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one schema method is provided (either block or JSON)
	hasSchemaBlock := len(data.Schema) > 0
	hasSchemaJSON := !data.SchemaJSON.IsNull() && !data.SchemaJSON.IsUnknown() && data.SchemaJSON.ValueString() != ""

	if hasSchemaBlock && hasSchemaJSON {
		resp.Diagnostics.AddError("Schema Error", "Cannot specify both schema block and schema_json attribute. Choose one.")
		return
	}

	if !hasSchemaBlock && !hasSchemaJSON {
		resp.Diagnostics.AddError("Schema Error", "Either schema block or schema_json attribute must be provided")
		return
	}

	var schemaMap map[string]any

	if hasSchemaBlock {
		// Validate that exactly one schema block is provided
		if len(data.Schema) != 1 {
			resp.Diagnostics.AddError("Schema Error", "Exactly one schema block must be provided")
			return
		}

		schemaBlock := data.Schema[0]

		// Build schema map from block attributes
		schemaMap = map[string]any{
			"title":       schemaBlock.Title.ValueString(),
			"description": schemaBlock.Description.ValueString(),
			"type":        schemaBlock.Type.ValueString(),
		}

		// Add properties if provided
		if !schemaBlock.Properties.IsNull() && !schemaBlock.Properties.IsUnknown() {
			var propertiesMap map[string]any
			if err := json.Unmarshal([]byte(schemaBlock.Properties.ValueString()), &propertiesMap); err != nil {
				resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to parse properties JSON: %s", err))
				return
			}
			schemaMap["properties"] = propertiesMap
		}

		// Add required fields if provided
		if !schemaBlock.Required.IsNull() && !schemaBlock.Required.IsUnknown() {
			var requiredList []string
			resp.Diagnostics.Append(schemaBlock.Required.ElementsAs(ctx, &requiredList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			schemaMap["required"] = requiredList
		}

		// Add strict if provided
		if !schemaBlock.Strict.IsNull() && !schemaBlock.Strict.IsUnknown() {
			schemaMap["strict"] = schemaBlock.Strict.ValueBool()
		}
	} else {
		// Parse schema JSON string
		if err := json.Unmarshal([]byte(data.SchemaJSON.ValueString()), &schemaMap); err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to parse schema JSON: %s", err))
			return
		}

		// Validate required fields in JSON schema
		if _, ok := schemaMap["title"]; !ok {
			resp.Diagnostics.AddError("Schema Error", "JSON schema must include 'title' field")
			return
		}
		if _, ok := schemaMap["description"]; !ok {
			resp.Diagnostics.AddError("Schema Error", "JSON schema must include 'description' field")
			return
		}
	}

	// Create class using the Tama client
	createRequest := neural.CreateClassRequest{
		Class: neural.ClassRequestData{
			Schema: schemaMap,
		},
	}

	tflog.Debug(ctx, "Creating class", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"schema":   schemaMap,
	})

	classResponse, err := r.client.Neural.CreateClass(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create class, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(classResponse.ID)
	data.Name = types.StringValue(classResponse.Name)
	data.Description = types.StringValue(classResponse.Description)
	data.ProvisionState = types.StringValue(classResponse.ProvisionState)
	data.SpaceId = types.StringValue(classResponse.SpaceID)

	// Update schema based on which method was used
	if hasSchemaBlock {
		err = r.updateSchemaFromResponse(ctx, classResponse.Schema, &data)
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to update schema from response: %s", err))
			return
		}
	} else {
		// Update schema_json with response, but normalize to match plan modifier behavior
		schemaJSON, err := json.Marshal(classResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to marshal schema to JSON: %s", err))
			return
		}

		// Normalize the marshaled JSON to ensure consistent formatting
		normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(schemaJSON))
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to normalize schema JSON: %s", err))
			return
		}
		data.SchemaJSON = types.StringValue(normalizedJSON)
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a class resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get class from API
	classResponse, err := r.client.Neural.GetClass(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read class, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Id = types.StringValue(classResponse.ID)
	data.Name = types.StringValue(classResponse.Name)
	data.Description = types.StringValue(classResponse.Description)
	data.ProvisionState = types.StringValue(classResponse.ProvisionState)
	data.SpaceId = types.StringValue(classResponse.SpaceID)

	// Update schema based on which method was used in current state
	hasSchemaBlock := len(data.Schema) > 0
	hasSchemaJSON := !data.SchemaJSON.IsNull() && !data.SchemaJSON.IsUnknown() && data.SchemaJSON.ValueString() != ""

	if hasSchemaBlock {
		err = r.updateSchemaFromResponse(ctx, classResponse.Schema, &data)
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to update schema from response: %s", err))
			return
		}
	} else if hasSchemaJSON {
		// Update schema_json with response, but normalize to match plan modifier behavior
		schemaJSON, err := json.Marshal(classResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to marshal schema to JSON: %s", err))
			return
		}

		// Normalize the marshaled JSON to ensure consistent formatting
		normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(schemaJSON))
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to normalize schema JSON: %s", err))
			return
		}
		data.SchemaJSON = types.StringValue(normalizedJSON)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one schema method is provided (either block or JSON)
	hasSchemaBlock := len(data.Schema) > 0
	hasSchemaJSON := !data.SchemaJSON.IsNull() && !data.SchemaJSON.IsUnknown() && data.SchemaJSON.ValueString() != ""

	if hasSchemaBlock && hasSchemaJSON {
		resp.Diagnostics.AddError("Schema Error", "Cannot specify both schema block and schema_json attribute. Choose one.")
		return
	}

	if !hasSchemaBlock && !hasSchemaJSON {
		resp.Diagnostics.AddError("Schema Error", "Either schema block or schema_json attribute must be provided")
		return
	}

	var schemaMap map[string]any

	if hasSchemaBlock {
		// Validate that exactly one schema block is provided
		if len(data.Schema) != 1 {
			resp.Diagnostics.AddError("Schema Error", "Exactly one schema block must be provided")
			return
		}

		schemaBlock := data.Schema[0]

		// Build schema map from block attributes
		schemaMap = map[string]any{
			"title":       schemaBlock.Title.ValueString(),
			"description": schemaBlock.Description.ValueString(),
			"type":        schemaBlock.Type.ValueString(),
		}

		// Add properties if provided
		if !schemaBlock.Properties.IsNull() && !schemaBlock.Properties.IsUnknown() {
			var propertiesMap map[string]any
			if err := json.Unmarshal([]byte(schemaBlock.Properties.ValueString()), &propertiesMap); err != nil {
				resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to parse properties JSON: %s", err))
				return
			}
			schemaMap["properties"] = propertiesMap
		}

		// Add required fields if provided
		if !schemaBlock.Required.IsNull() && !schemaBlock.Required.IsUnknown() {
			var requiredList []string
			resp.Diagnostics.Append(schemaBlock.Required.ElementsAs(ctx, &requiredList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			schemaMap["required"] = requiredList
		}

		// Add strict if provided
		if !schemaBlock.Strict.IsNull() && !schemaBlock.Strict.IsUnknown() {
			schemaMap["strict"] = schemaBlock.Strict.ValueBool()
		}
	} else {
		// Parse schema JSON string
		if err := json.Unmarshal([]byte(data.SchemaJSON.ValueString()), &schemaMap); err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to parse schema JSON: %s", err))
			return
		}

		// Validate required fields in JSON schema
		if _, ok := schemaMap["title"]; !ok {
			resp.Diagnostics.AddError("Schema Error", "JSON schema must include 'title' field")
			return
		}
		if _, ok := schemaMap["description"]; !ok {
			resp.Diagnostics.AddError("Schema Error", "JSON schema must include 'description' field")
			return
		}
	}

	// Update class using the Tama client
	updateRequest := neural.UpdateClassRequest{
		Class: neural.UpdateClassData{
			Schema: schemaMap,
		},
	}

	tflog.Debug(ctx, "Updating class", map[string]any{
		"id":     data.Id.ValueString(),
		"schema": schemaMap,
	})

	classResponse, err := r.client.Neural.UpdateClass(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update class, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Id = types.StringValue(classResponse.ID)
	data.Name = types.StringValue(classResponse.Name)
	data.Description = types.StringValue(classResponse.Description)
	data.ProvisionState = types.StringValue(classResponse.ProvisionState)
	data.SpaceId = types.StringValue(classResponse.SpaceID)

	// Update schema based on which method was used
	if hasSchemaBlock {
		err = r.updateSchemaFromResponse(ctx, classResponse.Schema, &data)
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to update schema from response: %s", err))
			return
		}
	} else {
		// Update schema_json with response, but normalize to match plan modifier behavior
		schemaJSON, err := json.Marshal(classResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to marshal schema to JSON: %s", err))
			return
		}

		// Normalize the marshaled JSON to ensure consistent formatting
		normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(schemaJSON))
		if err != nil {
			resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to normalize schema JSON: %s", err))
			return
		}
		data.SchemaJSON = types.StringValue(normalizedJSON)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete class using the Tama client
	tflog.Debug(ctx, "Deleting class", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteClass(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete class, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get class from API to populate state
	classResponse, err := r.client.Neural.GetClass(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import class, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:             types.StringValue(classResponse.ID),
		Name:           types.StringValue(classResponse.Name),
		Description:    types.StringValue(classResponse.Description),
		ProvisionState: types.StringValue(classResponse.ProvisionState),
		SpaceId:        types.StringValue(classResponse.SpaceID),
	}

	// For import, populate both schema formats to maintain compatibility
	// Update schema block with response data
	err = r.updateSchemaFromResponse(ctx, classResponse.Schema, &data)
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

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateSchemaFromResponse updates the schema block in the resource model from the API response.
func (r *Resource) updateSchemaFromResponse(ctx context.Context, responseSchema map[string]any, data *ResourceModel) error {
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
