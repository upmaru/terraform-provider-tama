// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package model

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
	"github.com/upmaru/tama-go/sensory"
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

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id         types.String `tfsdk:"id"`
	SourceId   types.String `tfsdk:"source_id"`
	Identifier types.String `tfsdk:"identifier"`
	Path       types.String `tfsdk:"path"`
	Parameters types.String `tfsdk:"parameters"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Sensory Model resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Model identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_id": schema.StringAttribute{
				MarkdownDescription: "ID of the source this model belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Model identifier (e.g., 'mistral-small-latest')",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "API path for the model (e.g., '/chat/completions')",
				Required:            true,
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "Model parameters as JSON string (e.g., '{\"temperature\": 0.8, \"max_tokens\": 1500}')",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					internalplanmodifier.JSONNormalize(),
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

	// Parse parameters if provided
	var parameters map[string]any
	if !data.Parameters.IsNull() && !data.Parameters.IsUnknown() && data.Parameters.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.Parameters.ValueString()), &parameters); err != nil {
			resp.Diagnostics.AddError("Invalid Parameters", fmt.Sprintf("Unable to parse parameters JSON: %s", err))
			return
		}
	}

	// Create model using the Tama client
	createRequest := sensory.CreateModelRequest{
		Model: sensory.ModelRequestData{
			Identifier: data.Identifier.ValueString(),
			Path:       data.Path.ValueString(),
			Parameters: parameters,
		},
	}

	tflog.Debug(ctx, "Creating model", map[string]any{
		"source_id":  data.SourceId.ValueString(),
		"identifier": data.Identifier.ValueString(),
		"path":       data.Path.ValueString(),
		"parameters": parameters,
	})

	modelResponse, err := r.client.Sensory.CreateModel(data.SourceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create model, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(modelResponse.ID)
	data.Identifier = types.StringValue(modelResponse.Identifier)
	// Note: Path is not returned in response, keep the original value

	// Handle parameters from response
	if len(modelResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(modelResponse.Parameters)
		if err != nil {
			resp.Diagnostics.AddError("Parameters Serialization Error", fmt.Sprintf("Unable to serialize parameters: %s", err))
			return
		}
		// Only update if user didn't provide parameters or if the values are different
		if data.Parameters.IsNull() || data.Parameters.IsUnknown() {
			data.Parameters = types.StringValue(string(parametersJSON))
		}
	} else if data.Parameters.IsNull() || data.Parameters.IsUnknown() {
		data.Parameters = types.StringValue("")
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a model resource")

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

	// Get model from API
	modelResponse, err := r.client.Sensory.GetModel(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read model, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Identifier = types.StringValue(modelResponse.Identifier)
	// Note: Path is not returned in response, keep the existing value

	// Handle parameters from response
	if len(modelResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(modelResponse.Parameters)
		if err != nil {
			resp.Diagnostics.AddError("Parameters Serialization Error", fmt.Sprintf("Unable to serialize parameters: %s", err))
			return
		}
		// Only update if the current value is null/unknown to preserve user input
		if data.Parameters.IsNull() || data.Parameters.IsUnknown() {
			data.Parameters = types.StringValue(string(parametersJSON))
		}
	} else if data.Parameters.IsNull() || data.Parameters.IsUnknown() {
		data.Parameters = types.StringValue("")
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

	// Parse parameters if provided
	var parameters map[string]any
	if !data.Parameters.IsNull() && !data.Parameters.IsUnknown() && data.Parameters.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.Parameters.ValueString()), &parameters); err != nil {
			resp.Diagnostics.AddError("Invalid Parameters", fmt.Sprintf("Unable to parse parameters JSON: %s", err))
			return
		}
	}

	// Update model using the Tama client
	updateRequest := sensory.UpdateModelRequest{
		Model: sensory.UpdateModelData{
			Identifier: data.Identifier.ValueString(),
			Path:       data.Path.ValueString(),
			Parameters: parameters,
		},
	}

	tflog.Debug(ctx, "Updating model", map[string]any{
		"id":         data.Id.ValueString(),
		"identifier": data.Identifier.ValueString(),
		"path":       data.Path.ValueString(),
		"parameters": parameters,
	})

	modelResponse, err := r.client.Sensory.UpdateModel(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update model, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Identifier = types.StringValue(modelResponse.Identifier)
	// Note: Path is not returned in response, keep the existing value

	// Handle parameters from response
	if len(modelResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(modelResponse.Parameters)
		if err != nil {
			resp.Diagnostics.AddError("Parameters Serialization Error", fmt.Sprintf("Unable to serialize parameters: %s", err))
			return
		}
		// Only update if the current value is null/unknown to preserve user input
		if data.Parameters.IsNull() || data.Parameters.IsUnknown() {
			data.Parameters = types.StringValue(string(parametersJSON))
		}
	} else if data.Parameters.IsNull() || data.Parameters.IsUnknown() {
		data.Parameters = types.StringValue("")
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

	// Delete model using the Tama client
	tflog.Debug(ctx, "Deleting model", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Sensory.DeleteModel(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete model, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get model from API to populate state
	modelResponse, err := r.client.Sensory.GetModel(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import model, got error: %s", err))
		return
	}

	// Handle parameters from response
	var parametersValue types.String
	if len(modelResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(modelResponse.Parameters)
		if err != nil {
			resp.Diagnostics.AddError("Parameters Serialization Error", fmt.Sprintf("Unable to serialize parameters: %s", err))
			return
		}
		parametersValue = types.StringValue(string(parametersJSON))
	} else {
		parametersValue = types.StringValue("")
	}

	// Create model from API response
	data := ResourceModel{
		Id:         types.StringValue(modelResponse.ID),
		Identifier: types.StringValue(modelResponse.Identifier),
		Parameters: parametersValue,
		// SourceId and Path cannot be retrieved from API response
		// These will need to be manually set after import
		SourceId: types.StringValue(""),
		Path:     types.StringValue(""),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
