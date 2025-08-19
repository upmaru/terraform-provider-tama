// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package initializer

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
	"github.com/upmaru/tama-go/tools"
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
	ThoughtToolId  types.String `tfsdk:"thought_tool_id"`
	Reference      types.String `tfsdk:"reference"`
	Index          types.Int64  `tfsdk:"index"`
	Parameters     types.String `tfsdk:"parameters"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_tool_initializer"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Tool Initializer resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Initializer identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_tool_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought tool this initializer belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reference": schema.StringAttribute{
				MarkdownDescription: "Reference path for the initializer. Valid references include 'tama/initializers/import' and 'tama/initializers/preload'",
				Required:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Index identifier for the initializer. If not provided, defaults to 0",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"parameters": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded parameters for the initializer",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					internalplanmodifier.JSONNormalize(),
				},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the initializer",
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

	// Set index if provided
	var index *int
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		indexValue := int(data.Index.ValueInt64())
		index = &indexValue
	}

	// Parse parameters if provided
	var parameters map[string]any
	if !data.Parameters.IsNull() && !data.Parameters.IsUnknown() && data.Parameters.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.Parameters.ValueString()), &parameters); err != nil {
			resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Invalid JSON in parameters: %s", err))
			return
		}
	}

	// Create initializer request
	createReq := tools.CreateInitializerRequest{
		Initializer: tools.InitializerRequestData{
			Reference:  data.Reference.ValueString(),
			Index:      index,
			Parameters: parameters,
		},
	}

	tflog.Debug(ctx, "Creating tool initializer", map[string]any{
		"thought_tool_id": data.ThoughtToolId.ValueString(),
		"reference":       createReq.Initializer.Reference,
		"index":           index,
	})

	// Create initializer
	initializerResponse, err := r.client.Tools.CreateInitializer(data.ThoughtToolId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tool initializer, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtToolId = types.StringValue(initializerResponse.ThoughtToolID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.Index = types.Int64Value(int64(initializerResponse.Index))
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

	// Handle parameters
	if err := r.updateParametersFromResponse(initializerResponse.Parameters, &data); err != nil {
		resp.Diagnostics.AddError("Parameters Error", fmt.Sprintf("Unable to update parameters from response: %s", err))
		return
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a tool initializer resource")

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

	// Get initializer from API
	tflog.Debug(ctx, "Reading tool initializer", map[string]any{
		"id": data.Id.ValueString(),
	})

	initializerResponse, err := r.client.Tools.GetInitializer(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool initializer, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtToolId = types.StringValue(initializerResponse.ThoughtToolID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.Index = types.Int64Value(int64(initializerResponse.Index))
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

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

	// Set index if provided
	var index *int
	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		indexValue := int(data.Index.ValueInt64())
		index = &indexValue
	}

	// Update initializer request
	updateReq := tools.UpdateInitializerRequest{
		Initializer: tools.UpdateInitializerData{
			Reference:  data.Reference.ValueString(),
			Index:      index,
			Parameters: parameters,
		},
	}

	tflog.Debug(ctx, "Updating tool initializer", map[string]any{
		"id":        data.Id.ValueString(),
		"reference": updateReq.Initializer.Reference,
		"index":     index,
	})

	// Update initializer
	initializerResponse, err := r.client.Tools.UpdateInitializer(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tool initializer, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtToolId = types.StringValue(initializerResponse.ThoughtToolID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.Index = types.Int64Value(int64(initializerResponse.Index))
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

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

	// Delete initializer
	tflog.Debug(ctx, "Deleting tool initializer", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Tools.DeleteInitializer(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tool initializer, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get initializer from API
	tflog.Debug(ctx, "Importing tool initializer", map[string]any{
		"id": req.ID,
	})

	initializerResponse, err := r.client.Tools.GetInitializer(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool initializer for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(initializerResponse.ID)
	data.ThoughtToolId = types.StringValue(initializerResponse.ThoughtToolID)
	data.Reference = types.StringValue(initializerResponse.Reference)
	data.Index = types.Int64Value(int64(initializerResponse.Index))
	data.ProvisionState = types.StringValue(initializerResponse.ProvisionState)

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
