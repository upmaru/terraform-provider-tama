// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package source_identity

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/sensory"
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

// ValidationModel describes the validation nested object.
type ValidationModel struct {
	Path   types.String `tfsdk:"path"`
	Method types.String `tfsdk:"method"`
	Codes  types.List   `tfsdk:"codes"`
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id              types.String     `tfsdk:"id"`
	SpecificationId types.String     `tfsdk:"specification_id"`
	Identifier      types.String     `tfsdk:"identifier"`
	ApiKey          types.String     `tfsdk:"api_key"`
	Validation      *ValidationModel `tfsdk:"validation"`
	ProvisionState  types.String     `tfsdk:"provision_state"`
	CurrentState    types.String     `tfsdk:"current_state"`
	WaitFor         []wait.WaitFor   `tfsdk:"wait_for"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_identity"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Sensory Source Identity resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identity identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"specification_id": schema.StringAttribute{
				MarkdownDescription: "ID of the specification this identity belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Identifier for the identity (e.g., 'ApiKey')",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for the identity",
				Required:            true,
				Sensitive:           true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current provision state of the identity",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the identity",
				Computed:            true,
			},
		},

		Blocks: func() map[string]schema.Block {
			blocks := map[string]schema.Block{
				"validation": schema.SingleNestedBlock{
					MarkdownDescription: "Validation configuration for the identity",
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Validation endpoint path",
							Required:            true,
						},
						"method": schema.StringAttribute{
							MarkdownDescription: "HTTP method for validation (e.g., 'GET', 'POST')",
							Required:            true,
						},
						"codes": schema.ListAttribute{
							MarkdownDescription: "List of acceptable HTTP status codes",
							Required:            true,
							ElementType:         types.Int64Type,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			}
			// Add wait_for blocks from the shared utility
			for key, block := range wait.WaitForBlockSchema() {
				blocks[key] = block
			}
			return blocks
		}(),
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

	// Convert codes from types.List to []int
	var codes []int64
	resp.Diagnostics.Append(data.Validation.Codes.ElementsAs(ctx, &codes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert int64 slice to int slice
	intCodes := make([]int, len(codes))
	for i, code := range codes {
		intCodes[i] = int(code)
	}

	// Create identity using the Tama client
	createRequest := sensory.CreateIdentityRequest{
		Identity: sensory.IdentityRequestData{
			APIKey: data.ApiKey.ValueString(),
			Validation: sensory.Validation{
				Path:   data.Validation.Path.ValueString(),
				Method: data.Validation.Method.ValueString(),
				Codes:  intCodes,
			},
		},
	}

	tflog.Debug(ctx, "Creating source identity", map[string]any{
		"specification_id":  data.SpecificationId.ValueString(),
		"identifier":        data.Identifier.ValueString(),
		"validation_path":   data.Validation.Path.ValueString(),
		"validation_method": data.Validation.Method.ValueString(),
	})

	identityResponse, err := r.client.Sensory.CreateIdentity(
		data.SpecificationId.ValueString(),
		data.Identifier.ValueString(),
		createRequest,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create source identity, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(identityResponse.ID)
	data.SpecificationId = types.StringValue(identityResponse.SpecificationID)
	data.Identifier = types.StringValue(identityResponse.Identifier)
	data.ProvisionState = types.StringValue(identityResponse.ProvisionState)
	data.CurrentState = types.StringValue(identityResponse.CurrentState)

	// Convert response validation codes back to types.List
	responseCodes := make([]int64, len(identityResponse.Validation.Codes))
	for i, code := range identityResponse.Validation.Codes {
		responseCodes[i] = int64(code)
	}
	codesList, diags := types.ListValueFrom(ctx, types.Int64Type, responseCodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Validation = &ValidationModel{
		Path:   types.StringValue(identityResponse.Validation.Path),
		Method: types.StringValue(identityResponse.Validation.Method),
		Codes:  codesList,
	}

	// Handle wait_for conditions if specified
	if len(data.WaitFor) > 0 {
		getIdentityFunc := func(id string) (interface{}, error) {
			return r.client.Sensory.GetIdentity(id)
		}
		for _, waitFor := range data.WaitFor {
			err := wait.ForConditions(ctx, getIdentityFunc, data.Id.ValueString(), waitFor.Field, 10*time.Minute)
			if err != nil {
				resp.Diagnostics.AddError("Wait Condition Failed", fmt.Sprintf("Unable to satisfy wait conditions: %s", err))
				return
			}
		}
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a source identity resource")

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

	// Get identity from API
	identityResponse, err := r.client.Sensory.GetIdentity(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read source identity, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.SpecificationId = types.StringValue(identityResponse.SpecificationID)
	data.Identifier = types.StringValue(identityResponse.Identifier)
	data.ProvisionState = types.StringValue(identityResponse.ProvisionState)
	data.CurrentState = types.StringValue(identityResponse.CurrentState)

	// Convert response validation codes to types.List
	responseCodes := make([]int64, len(identityResponse.Validation.Codes))
	for i, code := range identityResponse.Validation.Codes {
		responseCodes[i] = int64(code)
	}
	codesList, diags := types.ListValueFrom(ctx, types.Int64Type, responseCodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Validation = &ValidationModel{
		Path:   types.StringValue(identityResponse.Validation.Path),
		Method: types.StringValue(identityResponse.Validation.Method),
		Codes:  codesList,
	}

	// Note: API key is not returned in response, keep the original value

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

	// Convert codes from types.List to []int
	var codes []int64
	resp.Diagnostics.Append(data.Validation.Codes.ElementsAs(ctx, &codes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert int64 slice to int slice
	intCodes := make([]int, len(codes))
	for i, code := range codes {
		intCodes[i] = int(code)
	}

	// Update identity using the Tama client
	updateRequest := sensory.UpdateIdentityRequest{
		Identity: sensory.UpdateIdentityData{
			APIKey: data.ApiKey.ValueString(),
			Validation: &sensory.Validation{
				Path:   data.Validation.Path.ValueString(),
				Method: data.Validation.Method.ValueString(),
				Codes:  intCodes,
			},
		},
	}

	tflog.Debug(ctx, "Updating source identity", map[string]any{
		"id":                data.Id.ValueString(),
		"validation_path":   data.Validation.Path.ValueString(),
		"validation_method": data.Validation.Method.ValueString(),
	})

	identityResponse, err := r.client.Sensory.UpdateIdentity(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update source identity, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.SpecificationId = types.StringValue(identityResponse.SpecificationID)
	data.Identifier = types.StringValue(identityResponse.Identifier)
	data.ProvisionState = types.StringValue(identityResponse.ProvisionState)
	data.CurrentState = types.StringValue(identityResponse.CurrentState)

	// Convert response validation codes back to types.List
	responseCodes := make([]int64, len(identityResponse.Validation.Codes))
	for i, code := range identityResponse.Validation.Codes {
		responseCodes[i] = int64(code)
	}
	codesList, diags := types.ListValueFrom(ctx, types.Int64Type, responseCodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Validation = &ValidationModel{
		Path:   types.StringValue(identityResponse.Validation.Path),
		Method: types.StringValue(identityResponse.Validation.Method),
		Codes:  codesList,
	}

	// Note: API key is not returned in response, keep the original value

	// Handle wait_for conditions if specified
	if len(data.WaitFor) > 0 {
		getIdentityFunc := func(id string) (interface{}, error) {
			return r.client.Sensory.GetIdentity(id)
		}
		for _, waitFor := range data.WaitFor {
			err := wait.ForConditions(ctx, getIdentityFunc, data.Id.ValueString(), waitFor.Field, 10*time.Minute)
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

	// Delete identity using the Tama client
	tflog.Debug(ctx, "Deleting source identity", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Sensory.DeleteIdentity(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete source identity, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get identity from API to populate state
	identityResponse, err := r.client.Sensory.GetIdentity(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import source identity, got error: %s", err))
		return
	}

	// Convert response validation codes to types.List
	responseCodes := make([]int64, len(identityResponse.Validation.Codes))
	for i, code := range identityResponse.Validation.Codes {
		responseCodes[i] = int64(code)
	}
	codesList, diags := types.ListValueFrom(ctx, types.Int64Type, responseCodes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:              types.StringValue(identityResponse.ID),
		SpecificationId: types.StringValue(identityResponse.SpecificationID),
		Identifier:      types.StringValue(identityResponse.Identifier),
		ProvisionState:  types.StringValue(identityResponse.ProvisionState),
		CurrentState:    types.StringValue(identityResponse.CurrentState),
		Validation: &ValidationModel{
			Path:   types.StringValue(identityResponse.Validation.Path),
			Method: types.StringValue(identityResponse.Validation.Method),
			Codes:  codesList,
		},
		// ApiKey cannot be retrieved from API response
		// This will need to be manually set after import
		ApiKey: types.StringValue(""),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
