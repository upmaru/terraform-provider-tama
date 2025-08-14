// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package thought_initializer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	Id             types.String `tfsdk:"id"`
	ThoughtId      types.String `tfsdk:"thought_id"`
	Reference      types.String `tfsdk:"reference"`
	Index          types.Int64  `tfsdk:"index"`
	ClassId        types.String `tfsdk:"class_id"`
	Parameters     types.String `tfsdk:"parameters"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_initializer"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Perception Thought Initializer resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Thought initializer identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this initializer belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reference": schema.StringAttribute{
				MarkdownDescription: "Initializer reference (e.g., 'tama/initializers/preload')",
				Required:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Index position of the initializer",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this initializer operates on",
				Required:            true,
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "Initializer parameters as JSON string",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					internalplanmodifier.JSONNormalize(),
				},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the thought initializer",
				Computed:            true,
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
			resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Invalid JSON in parameters: %s", err))
			return
		}
	}

	// Create thought initializer request
	createReq := perception.CreateInitializerRequest{
		Initializer: perception.InitializerRequestData{
			ClassID:    data.ClassId.ValueString(),
			Reference:  data.Reference.ValueString(),
			Parameters: parameters,
		},
	}

	// Add index if provided and not empty
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		index := int(data.Index.ValueInt64())
		createReq.Initializer.Index = &index
	}

	tflog.Debug(ctx, "Creating thought initializer", map[string]any{
		"thought_id": data.ThoughtId.ValueString(),
		"reference":  createReq.Initializer.Reference,
		"class_id":   createReq.Initializer.ClassID,
	})

	// Create thought initializer
	initializerResponse, err := r.client.Perception.CreateInitializer(data.ThoughtId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create thought initializer, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtId = types.StringValue(initializerResponse.ThoughtID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.ClassId = types.StringValue(initializerResponse.ClassID)
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

	// Handle index
	if initializerResponse.Index != nil {
		data.Index = types.Int64Value(int64(*initializerResponse.Index))
	} else {
		data.Index = types.Int64Null()
	}

	// Handle parameters
	if err := r.updateParametersFromResponse(initializerResponse.Parameters, &data); err != nil {
		resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Unable to update parameters from response: %s", err))
		return
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a thought initializer resource")

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

	// Get thought initializer from API
	tflog.Debug(ctx, "Reading thought initializer", map[string]any{
		"id": data.Id.ValueString(),
	})

	initializerResponse, err := r.client.Perception.GetInitializer(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought initializer, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtId = types.StringValue(initializerResponse.ThoughtID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.ClassId = types.StringValue(initializerResponse.ClassID)
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

	// Handle index
	if initializerResponse.Index != nil {
		data.Index = types.Int64Value(int64(*initializerResponse.Index))
	} else {
		data.Index = types.Int64Null()
	}

	// Handle parameters
	if err := r.updateParametersFromResponse(initializerResponse.Parameters, &data); err != nil {
		resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Unable to update parameters from response: %s", err))
		return
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
			resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Invalid JSON in parameters: %s", err))
			return
		}
	}

	// Update thought initializer request
	updateReq := perception.UpdateInitializerRequest{
		Initializer: perception.UpdateInitializerData{
			ClassID:    data.ClassId.ValueString(),
			Reference:  data.Reference.ValueString(),
			Parameters: parameters,
		},
	}

	// Add index if provided and not empty
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		index := int(data.Index.ValueInt64())
		updateReq.Initializer.Index = &index
	}

	tflog.Debug(ctx, "Updating thought initializer", map[string]any{
		"id":        data.Id.ValueString(),
		"reference": updateReq.Initializer.Reference,
		"class_id":  updateReq.Initializer.ClassID,
	})

	// Update thought initializer
	initializerResponse, err := r.client.Perception.UpdateInitializer(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update thought initializer, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtId = types.StringValue(initializerResponse.ThoughtID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.ClassId = types.StringValue(initializerResponse.ClassID)
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

	// Handle index
	if initializerResponse.Index != nil {
		data.Index = types.Int64Value(int64(*initializerResponse.Index))
	} else {
		data.Index = types.Int64Null()
	}

	// Handle parameters
	if err := r.updateParametersFromResponse(initializerResponse.Parameters, &data); err != nil {
		resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Unable to update parameters from response: %s", err))
		return
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

	// Delete thought initializer
	tflog.Debug(ctx, "Deleting thought initializer", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteInitializer(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete thought initializer, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get thought initializer from API
	tflog.Debug(ctx, "Importing thought initializer", map[string]any{
		"id": req.ID,
	})

	initializerResponse, err := r.client.Perception.GetInitializer(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought initializer for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtId = types.StringValue(initializerResponse.ThoughtID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.ClassId = types.StringValue(initializerResponse.ClassID)
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

	// Handle index
	if initializerResponse.Index != nil {
		data.Index = types.Int64Value(int64(*initializerResponse.Index))
	} else {
		data.Index = types.Int64Null()
	}

	// Handle parameters
	if err := r.updateParametersFromResponse(initializerResponse.Parameters, &data); err != nil {
		resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Unable to update parameters from response: %s", err))
		return
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateParametersFromResponse updates the parameters field in the resource model from the API response.
func (r *Resource) updateParametersFromResponse(responseParameters map[string]any, data *ResourceModel) error {
	// Handle parameters - always use the server response as the source of truth
	if responseParameters != nil {
		// Use server response as-is since it includes defaults and server-side processing
		parametersJSON, err := json.Marshal(responseParameters)
		if err != nil {
			return fmt.Errorf("unable to marshal parameters: %s", err)
		}

		// Normalize the marshaled JSON to ensure consistent formatting
		normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(parametersJSON))
		if err != nil {
			return fmt.Errorf("unable to normalize parameters JSON: %s", err)
		}
		data.Parameters = types.StringValue(normalizedJSON)
	} else {
		data.Parameters = types.StringNull()
	}

	return nil
}

// preserveUserFloatTypes merges server response parameters with user-provided parameters,
// preserving the user's original float types when the server converts them to strings.
func (r *Resource) preserveUserFloatTypes(userParams, serverParams map[string]any) map[string]any {
	result := make(map[string]any)

	// Start with server parameters (includes any new parameters the server added)
	for k, v := range serverParams {
		result[k] = v
	}

	// Override with user parameters, preserving their original float types
	for k, userValue := range userParams {
		serverValue, exists := serverParams[k]
		if !exists {
			// User parameter doesn't exist in server response, keep user value
			result[k] = userValue
			continue
		}

		// Handle nested objects recursively
		if userMap, userIsMap := userValue.(map[string]any); userIsMap {
			if serverMap, serverIsMap := serverValue.(map[string]any); serverIsMap {
				// Both are maps, merge recursively
				result[k] = r.preserveUserFloatTypes(userMap, serverMap)
				continue
			}
		}

		// Preserve user's float types when server converts them to strings
		if userFloat, userIsFloat := userValue.(float64); userIsFloat {
			if serverStr, serverIsString := serverValue.(string); serverIsString {
				// Check if the string representation matches the float
				if fmt.Sprintf("%g", userFloat) == serverStr {
					result[k] = userValue // Preserve the original float
					continue
				}
			}
		}

		// For other cases, prefer server value (it might have been updated)
		result[k] = serverValue
	}

	return result
}
