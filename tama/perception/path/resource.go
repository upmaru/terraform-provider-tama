// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path

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
	"github.com/upmaru/tama-go/perception"
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
	Id            types.String `tfsdk:"id"`
	ThoughtId     types.String `tfsdk:"thought_id"`
	TargetClassId types.String `tfsdk:"target_class_id"`
	Parameters    types.String `tfsdk:"parameters"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_path"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Path resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Path identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this path belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the target class for this path",
				Required:            true,
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "Path parameters as JSON string (e.g., '{\"similarity\": {\"threshold\": 0.9}}')",
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

	// Create path using the Tama client
	createRequest := perception.CreatePathRequest{
		Path: perception.PathRequestData{
			TargetClassID: data.TargetClassId.ValueString(),
			Parameters:    parameters,
		},
	}

	tflog.Debug(ctx, "Creating path", map[string]any{
		"thought_id":      data.ThoughtId.ValueString(),
		"target_class_id": data.TargetClassId.ValueString(),
		"parameters":      parameters,
	})

	pathResponse, err := r.client.Perception.CreatePath(data.ThoughtId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create path, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(pathResponse.ID)
	data.TargetClassId = types.StringValue(pathResponse.TargetClassID)

	// Handle parameters from response
	if len(pathResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(pathResponse.Parameters)
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
	tflog.Trace(ctx, "created a path resource")

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

	// Get path from API
	pathResponse, err := r.client.Perception.GetPath(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read path, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.TargetClassId = types.StringValue(pathResponse.TargetClassID)

	// Handle parameters from response
	if len(pathResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(pathResponse.Parameters)
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

	// Update path using the Tama client
	updateRequest := perception.UpdatePathRequest{
		Path: perception.UpdatePathData{
			TargetClassID: data.TargetClassId.ValueString(),
			Parameters:    parameters,
		},
	}

	tflog.Debug(ctx, "Updating path", map[string]any{
		"id":              data.Id.ValueString(),
		"target_class_id": data.TargetClassId.ValueString(),
		"parameters":      parameters,
	})

	pathResponse, err := r.client.Perception.UpdatePath(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update path, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.TargetClassId = types.StringValue(pathResponse.TargetClassID)

	// Handle parameters from response
	if len(pathResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(pathResponse.Parameters)
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

	// Delete path using the Tama client
	tflog.Debug(ctx, "Deleting path", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeletePath(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete path, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get path from API to populate state
	pathResponse, err := r.client.Perception.GetPath(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import path, got error: %s", err))
		return
	}

	// Handle parameters from response
	var parametersValue types.String
	if len(pathResponse.Parameters) > 0 {
		parametersJSON, err := json.Marshal(pathResponse.Parameters)
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
		Id:            types.StringValue(pathResponse.ID),
		TargetClassId: types.StringValue(pathResponse.TargetClassID),
		Parameters:    parametersValue,
		// ThoughtId cannot be retrieved from API response
		// This will need to be manually set after import
		ThoughtId: types.StringValue(pathResponse.ThoughtID),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
