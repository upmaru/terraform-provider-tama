// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space_processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/neural"
	"github.com/upmaru/terraform-provider-tama/internal/processor"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithConfigValidators = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *tama.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_processor"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes, blocks := processor.GetNeuralProcessorSchema()
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Space Processor resource",
		Attributes:          attributes,
		Blocks:              blocks,
	}
}

func (r *Resource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("completion"),
			path.MatchRoot("embedding"),
			path.MatchRoot("reranking"),
		),
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
	var data processor.NeuralProcessorModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine type and validate configuration
	processorType := processor.DetermineProcessorType(&data)
	if processorType == "" {
		resp.Diagnostics.AddError("Configuration Error", "exactly one configuration block must be provided (completion, embedding, or reranking)")
		return
	}

	// Set the type in the data model
	data.Type = types.StringValue(processorType)

	// Build configuration based on type
	config := processor.BuildConfiguration(&data)

	// Create processor using the Tama client
	createRequest := neural.CreateProcessorRequest{
		Processor: neural.ProcessorRequestData{
			ModelID:       data.ModelId.ValueString(),
			Configuration: config,
		},
	}

	tflog.Debug(ctx, "Creating processor", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"model_id": data.ModelId.ValueString(),
		"type":     processorType,
		"config":   config,
	})

	processorResponse, err := r.client.Neural.CreateProcessor(data.SpaceId.ValueString(), processorType, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create processor, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(processorResponse.ID)
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)

	// Ensure parameters are initialized to avoid unknown state
	processor.EnsureParametersInitialized(&data)

	// Update configuration blocks based on the type and API response
	processor.UpdateConfigurationFromResponse(processorResponse.Configuration, &data)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a processor resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data processor.NeuralProcessorModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get processor from API
	processorResponse, err := r.client.Neural.GetProcessor(data.SpaceId.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read processor, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)

	// Update configuration blocks based on the type and API response
	processor.UpdateConfigurationFromResponse(processorResponse.Configuration, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data processor.NeuralProcessorModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine type and validate configuration
	processorType := processor.DetermineProcessorType(&data)
	if processorType == "" {
		resp.Diagnostics.AddError("Configuration Error", "exactly one configuration block must be provided (completion, embedding, or reranking)")
		return
	}

	// Set the type in the data model
	data.Type = types.StringValue(processorType)

	// Ensure parameters are initialized to avoid unknown state
	processor.EnsureParametersInitialized(&data)

	// Build configuration based on type
	config := processor.BuildConfiguration(&data)

	// Update processor using the Tama client
	updateRequest := neural.UpdateProcessorRequest{
		Processor: neural.UpdateProcessorData{
			ModelID:       data.ModelId.ValueString(),
			Configuration: config,
		},
	}

	tflog.Debug(ctx, "Updating processor", map[string]any{
		"id":       data.Id.ValueString(),
		"model_id": data.ModelId.ValueString(),
		"type":     processorType,
		"config":   config,
	})

	processorResponse, err := r.client.Neural.UpdateProcessor(data.SpaceId.ValueString(), processorType, updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update processor, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.ModelId = types.StringValue(processorResponse.ModelID)
	data.Type = types.StringValue(processorResponse.Type)

	// Update configuration blocks based on the type and API response
	processor.UpdateConfigurationFromResponse(processorResponse.Configuration, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data processor.NeuralProcessorModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete processor using the Tama client
	tflog.Debug(ctx, "Deleting processor", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteProcessor(data.SpaceId.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete processor, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the compound ID to extract space_id and type
	// The import ID should be in the format "space_id/type"
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'space_id/type'",
		)
		return
	}

	spaceID := parts[0]
	processorType := parts[1]

	// Validate processor type
	validTypes := []string{"completion", "embedding", "reranking"}
	isValidType := false
	for _, validType := range validTypes {
		if processorType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		resp.Diagnostics.AddError(
			"Invalid Processor Type",
			fmt.Sprintf("Processor type must be one of: %v", validTypes),
		)
		return
	}

	// Get processor from API to populate state
	processorResponse, err := r.client.Neural.GetProcessor(spaceID, processorType)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import processor, got error: %s", err))
		return
	}

	// Create model from API response using shared model
	data := processor.NeuralProcessorModel{
		SpaceId: types.StringValue(spaceID),
		ProcessorModel: processor.ProcessorModel{
			Id:      types.StringValue(processorResponse.ID),
			ModelId: types.StringValue(processorResponse.ModelID),
			Type:    types.StringValue(processorResponse.Type),
		},
	}

	// Update configuration blocks based on the type and API response
	processor.UpdateConfigurationFromResponseWithType(processorResponse.Configuration, &data, processorType)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
