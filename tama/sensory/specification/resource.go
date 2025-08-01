// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package specification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/sensory"
	internalplanmodifier "github.com/upmaru/terraform-provider-tama/internal/planmodifier"
	"github.com/upmaru/terraform-provider-tama/internal/wait"
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
	Id             types.String   `tfsdk:"id"`
	SpaceId        types.String   `tfsdk:"space_id"`
	Schema         types.String   `tfsdk:"schema"`
	Version        types.String   `tfsdk:"version"`
	Endpoint       types.String   `tfsdk:"endpoint"`
	CurrentState   types.String   `tfsdk:"current_state"`
	ProvisionState types.String   `tfsdk:"provision_state"`
	WaitFor        []wait.WaitFor `tfsdk:"wait_for"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_specification"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Sensory Specification resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Specification identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this specification belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "OpenAPI 3.0 schema definition for the specification",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					internalplanmodifier.JSONNormalize(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the specification",
				Required:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint URL for the specification",
				Required:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the specification",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Provision state of the specification",
				Computed:            true,
			},
		},
		Blocks: wait.WaitForBlockSchema(),
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

	// Parse schema JSON
	var schemaMap map[string]any
	if err := json.Unmarshal([]byte(data.Schema.ValueString()), &schemaMap); err != nil {
		resp.Diagnostics.AddError("Invalid Schema", fmt.Sprintf("Unable to parse schema JSON: %s", err))
		return
	}

	// Create specification using the Tama client
	createRequest := sensory.CreateSpecificationRequest{
		Specification: sensory.SpecificationRequestData{
			Schema:   schemaMap,
			Version:  data.Version.ValueString(),
			Endpoint: data.Endpoint.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating specification", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"version":  data.Version.ValueString(),
		"endpoint": data.Endpoint.ValueString(),
		"schema":   schemaMap,
	})

	specResponse, err := r.client.Sensory.CreateSpecification(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create specification, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(specResponse.ID)
	data.SpaceId = types.StringValue(specResponse.SpaceID)
	data.Version = types.StringValue(specResponse.Version)
	data.Endpoint = types.StringValue(specResponse.Endpoint)
	data.CurrentState = types.StringValue(specResponse.CurrentState)
	data.ProvisionState = types.StringValue(specResponse.ProvisionState)

	// Handle schema from response
	if len(specResponse.Schema) > 0 {
		schemaJSON, err := json.Marshal(specResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Serialization Error", fmt.Sprintf("Unable to serialize schema: %s", err))
			return
		}
		data.Schema = types.StringValue(string(schemaJSON))
	}

	// Handle wait_for conditions if specified
	if len(data.WaitFor) > 0 {
		getSpecificationFunc := func(id string) (interface{}, error) {
			return r.client.Sensory.GetSpecification(id)
		}
		for _, waitFor := range data.WaitFor {
			err := wait.ForConditions(ctx, getSpecificationFunc, data.Id.ValueString(), waitFor.Field, 10*time.Minute)
			if err != nil {
				resp.Diagnostics.AddError("Wait Condition Failed", fmt.Sprintf("Unable to satisfy wait conditions: %s", err))
				return
			}
		}
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a specification resource")

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

	// Get specification from API
	specResponse, err := r.client.Sensory.GetSpecification(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read specification, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.SpaceId = types.StringValue(specResponse.SpaceID)
	data.Version = types.StringValue(specResponse.Version)
	data.Endpoint = types.StringValue(specResponse.Endpoint)
	data.CurrentState = types.StringValue(specResponse.CurrentState)
	data.ProvisionState = types.StringValue(specResponse.ProvisionState)

	// Handle schema from response
	if len(specResponse.Schema) > 0 {
		schemaJSON, err := json.Marshal(specResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Serialization Error", fmt.Sprintf("Unable to serialize schema: %s", err))
			return
		}
		data.Schema = types.StringValue(string(schemaJSON))
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

	// Parse schema JSON
	var schemaMap map[string]any
	if err := json.Unmarshal([]byte(data.Schema.ValueString()), &schemaMap); err != nil {
		resp.Diagnostics.AddError("Invalid Schema", fmt.Sprintf("Unable to parse schema JSON: %s", err))
		return
	}

	// Update specification using the Tama client
	updateRequest := sensory.UpdateSpecificationRequest{
		Specification: sensory.UpdateSpecificationData{
			Schema:   schemaMap,
			Version:  data.Version.ValueString(),
			Endpoint: data.Endpoint.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating specification", map[string]any{
		"id":       data.Id.ValueString(),
		"version":  data.Version.ValueString(),
		"endpoint": data.Endpoint.ValueString(),
		"schema":   schemaMap,
	})

	specResponse, err := r.client.Sensory.UpdateSpecification(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update specification, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.SpaceId = types.StringValue(specResponse.SpaceID)
	data.Version = types.StringValue(specResponse.Version)
	data.Endpoint = types.StringValue(specResponse.Endpoint)
	data.CurrentState = types.StringValue(specResponse.CurrentState)
	data.ProvisionState = types.StringValue(specResponse.ProvisionState)

	// Handle schema from response
	if len(specResponse.Schema) > 0 {
		schemaJSON, err := json.Marshal(specResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Serialization Error", fmt.Sprintf("Unable to serialize schema: %s", err))
			return
		}
		data.Schema = types.StringValue(string(schemaJSON))
	}

	// Handle wait_for conditions if specified
	if len(data.WaitFor) > 0 {
		getSpecificationFunc := func(id string) (interface{}, error) {
			return r.client.Sensory.GetSpecification(id)
		}
		for _, waitFor := range data.WaitFor {
			err := wait.ForConditions(ctx, getSpecificationFunc, data.Id.ValueString(), waitFor.Field, 10*time.Minute)
			if err != nil {
				resp.Diagnostics.AddError("Wait Condition Failed", fmt.Sprintf("Unable to satisfy wait conditions: %s", err))
				return
			}
		}
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

	// Delete specification using the Tama client
	tflog.Debug(ctx, "Deleting specification", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Sensory.DeleteSpecification(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete specification, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get specification from API to populate state
	specResponse, err := r.client.Sensory.GetSpecification(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import specification, got error: %s", err))
		return
	}

	// Handle schema from response
	var schemaValue types.String
	if len(specResponse.Schema) > 0 {
		schemaJSON, err := json.Marshal(specResponse.Schema)
		if err != nil {
			resp.Diagnostics.AddError("Schema Serialization Error", fmt.Sprintf("Unable to serialize schema: %s", err))
			return
		}
		schemaValue = types.StringValue(string(schemaJSON))
	} else {
		schemaValue = types.StringValue("")
	}

	// Create model from API response
	data := ResourceModel{
		Id:             types.StringValue(specResponse.ID),
		SpaceId:        types.StringValue(specResponse.SpaceID),
		Schema:         schemaValue,
		Version:        types.StringValue(specResponse.Version),
		Endpoint:       types.StringValue(specResponse.Endpoint),
		CurrentState:   types.StringValue(specResponse.CurrentState),
		ProvisionState: types.StringValue(specResponse.ProvisionState),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
