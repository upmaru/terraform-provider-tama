// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package modular_thought

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

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

// ModuleModel describes the module block data model.
type ModuleModel struct {
	Reference  types.String `tfsdk:"reference"`
	Parameters types.String `tfsdk:"parameters"`
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id             types.String `tfsdk:"id"`
	ChainId        types.String `tfsdk:"chain_id"`
	OutputClassId  types.String `tfsdk:"output_class_id"`
	Module         ModuleModel  `tfsdk:"module"`
	ProvisionState types.String `tfsdk:"provision_state"`
	Relation       types.String `tfsdk:"relation"`
	Index          types.Int64  `tfsdk:"index"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_modular_thought"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Perception Modular Thought resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Modular thought identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain this modular thought belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"output_class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the output class for this modular thought",
				Optional:            true,
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the modular thought",
				Computed:            true,
			},
			"relation": schema.StringAttribute{
				MarkdownDescription: "Relation type for the modular thought (e.g., 'description', 'analysis')",
				Required:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Index position of the modular thought in the chain",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"module": schema.SingleNestedBlock{
				MarkdownDescription: "Module configuration for the modular thought",
				Attributes: map[string]schema.Attribute{
					"reference": schema.StringAttribute{
						MarkdownDescription: "Module reference (e.g., 'tama/agentic/generate')",
						Required:            true,
					},
					"parameters": schema.StringAttribute{
						MarkdownDescription: "Module parameters as JSON string",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							internalplanmodifier.JSONNormalize(),
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

	moduleBlock := data.Module

	// Parse module parameters if provided
	var parameters map[string]any
	if !moduleBlock.Parameters.IsNull() && !moduleBlock.Parameters.IsUnknown() && moduleBlock.Parameters.ValueString() != "" {
		if err := json.Unmarshal([]byte(moduleBlock.Parameters.ValueString()), &parameters); err != nil {
			resp.Diagnostics.AddError("Module Parameters Error", fmt.Sprintf("Invalid JSON in module parameters: %s", err))
			return
		}
	}

	// Create modular thought request
	createReq := perception.CreateThoughtRequest{
		Thought: perception.ThoughtRequestData{
			Relation: data.Relation.ValueString(),
			Module: &perception.Module{
				Reference:  moduleBlock.Reference.ValueString(),
				Parameters: parameters,
			},
		},
	}

	// Add output_class_id if provided and not empty
	if !data.OutputClassId.IsNull() && !data.OutputClassId.IsUnknown() && data.OutputClassId.ValueString() != "" {
		createReq.Thought.OutputClassID = data.OutputClassId.ValueString()
	}

	// Add index if provided and not empty
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		index := int(data.Index.ValueInt64())
		createReq.Thought.Index = &index
	}

	tflog.Debug(ctx, "Creating modular thought", map[string]any{
		"chain_id":         data.ChainId.ValueString(),
		"relation":         createReq.Thought.Relation,
		"module_reference": createReq.Thought.Module.Reference,
	})

	// Create modular thought
	thoughtResponse, err := r.client.Perception.CreateThought(data.ChainId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create modular thought, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update module with response data
	err = r.updateModuleFromResponse(*thoughtResponse.Module, &data)
	if err != nil {
		resp.Diagnostics.AddError("Module Error", fmt.Sprintf("Unable to update module from response: %s", err))
		return
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a modular thought resource")

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

	// Get modular thought from API
	tflog.Debug(ctx, "Reading modular thought", map[string]any{
		"id": data.Id.ValueString(),
	})

	thoughtResponse, err := r.client.Perception.GetThought(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read modular thought for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update module with response data
	err = r.updateModuleFromResponse(*thoughtResponse.Module, &data)
	if err != nil {
		resp.Diagnostics.AddError("Module Error", fmt.Sprintf("Unable to update module from response: %s", err))
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

	moduleBlock := data.Module

	// Parse module parameters if provided
	var parameters map[string]any
	if !moduleBlock.Parameters.IsNull() && !moduleBlock.Parameters.IsUnknown() && moduleBlock.Parameters.ValueString() != "" {
		if err := json.Unmarshal([]byte(moduleBlock.Parameters.ValueString()), &parameters); err != nil {
			resp.Diagnostics.AddError("Module Parameters Error", fmt.Sprintf("Invalid JSON in module parameters: %s", err))
			return
		}
	}

	// Update modular thought request
	updateReq := perception.UpdateThoughtRequest{
		Thought: perception.UpdateThoughtData{
			Relation: data.Relation.ValueString(),
			Module: &perception.Module{
				Reference:  moduleBlock.Reference.ValueString(),
				Parameters: parameters,
			},
		},
	}

	// Add output_class_id if provided and not empty
	if !data.OutputClassId.IsNull() && !data.OutputClassId.IsUnknown() && data.OutputClassId.ValueString() != "" {
		updateReq.Thought.OutputClassID = data.OutputClassId.ValueString()
	}

	// Add index if provided and not empty
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		index := int(data.Index.ValueInt64())
		updateReq.Thought.Index = &index
	}

	tflog.Debug(ctx, "Updating modular thought", map[string]any{
		"id":               data.Id.ValueString(),
		"relation":         updateReq.Thought.Relation,
		"module_reference": updateReq.Thought.Module.Reference,
	})

	// Update modular thought
	thoughtResponse, err := r.client.Perception.UpdateThought(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update modular thought, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update module with response data
	err = r.updateModuleFromResponse(*thoughtResponse.Module, &data)
	if err != nil {
		resp.Diagnostics.AddError("Module Error", fmt.Sprintf("Unable to update module from response: %s", err))
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

	// Delete thought
	tflog.Debug(ctx, "Deleting modular thought", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteThought(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete modular thought, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get thought from API
	tflog.Debug(ctx, "Importing thought", map[string]any{
		"id": req.ID,
	})

	thoughtResponse, err := r.client.Perception.GetThought(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update module with response data
	err = r.updateModuleFromResponse(*thoughtResponse.Module, &data)
	if err != nil {
		resp.Diagnostics.AddError("Module Error", fmt.Sprintf("Unable to update module from response: %s", err))
		return
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModuleFromResponse updates the module block in the resource model from the API response.
// It attempts to preserve user-provided float types when the server converts them to strings.
func (r *Resource) updateModuleFromResponse(responseModule perception.Module, data *ResourceModel) error {
	moduleModel := ModuleModel{
		Reference: types.StringValue(responseModule.Reference),
	}

	// Handle parameters
	if responseModule.Parameters != nil {
		// If we have existing module data, try to preserve user types
		if !data.Module.Parameters.IsNull() && !data.Module.Parameters.IsUnknown() {
			existingParamsStr := data.Module.Parameters.ValueString()
			if existingParamsStr != "" {
				// Parse existing parameters to get original types
				var existingParams map[string]any
				if err := json.Unmarshal([]byte(existingParamsStr), &existingParams); err == nil {
					// Merge response parameters with existing ones, preserving float types
					mergedParams := preserveUserFloatTypes(existingParams, responseModule.Parameters)

					// Use merged parameters
					parametersJSON, err := json.Marshal(mergedParams)
					if err != nil {
						return fmt.Errorf("unable to marshal merged module parameters: %s", err)
					}

					// Normalize the marshaled JSON to ensure consistent formatting
					normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(parametersJSON))
					if err != nil {
						return fmt.Errorf("unable to normalize merged module parameters JSON: %s", err)
					}
					moduleModel.Parameters = types.StringValue(normalizedJSON)

					data.Module = moduleModel
					return nil
				}
			}
		}

		// Fallback: use server response as-is
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

	data.Module = moduleModel
	return nil
}

// preserveUserFloatTypes merges server response parameters with user-provided parameters,
// preserving the user's original float types when the server converts them to strings.
func preserveUserFloatTypes(userParams, serverParams map[string]any) map[string]any {
	result := make(map[string]any)

	// Start with server parameters (includes any new parameters the server added)
	maps.Copy(result, serverParams)

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
				result[k] = preserveUserFloatTypes(userMap, serverMap)
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
