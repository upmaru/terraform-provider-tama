// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package activation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/perception"
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
	ThoughtPathId  types.String `tfsdk:"thought_path_id"`
	ChainId        types.String `tfsdk:"chain_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_path_activation"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Path Activation resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Activation identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_path_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought path this activation belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain to activate for this path",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the activation",
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

	// Create activation request
	createReq := perception.CreateActivationRequest{
		Activation: perception.ActivationRequestData{
			ChainID: data.ChainId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating activation", map[string]any{
		"thought_path_id": data.ThoughtPathId.ValueString(),
		"chain_id":        createReq.Activation.ChainID,
	})

	// Create activation
	activationResponse, err := r.client.Perception.CreateActivation(data.ThoughtPathId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create activation, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(activationResponse.ID)
	data.ThoughtPathId = types.StringValue(activationResponse.ThoughtPathID)
	data.ChainId = types.StringValue(activationResponse.ChainID)
	data.ProvisionState = types.StringValue(activationResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created an activation resource")

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

	// Get activation from API
	tflog.Debug(ctx, "Reading activation", map[string]any{
		"id": data.Id.ValueString(),
	})

	activationResponse, err := r.client.Perception.GetActivation(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read activation, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(activationResponse.ID)
	data.ThoughtPathId = types.StringValue(activationResponse.ThoughtPathID)
	data.ChainId = types.StringValue(activationResponse.ChainID)
	data.ProvisionState = types.StringValue(activationResponse.ProvisionState)

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

	// Update activation request
	updateReq := perception.UpdateActivationRequest{
		Activation: perception.UpdateActivationData{
			ChainID: data.ChainId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating activation", map[string]any{
		"id":       data.Id.ValueString(),
		"chain_id": updateReq.Activation.ChainID,
	})

	// Update activation
	activationResponse, err := r.client.Perception.UpdateActivation(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update activation, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(activationResponse.ID)
	data.ThoughtPathId = types.StringValue(activationResponse.ThoughtPathID)
	data.ChainId = types.StringValue(activationResponse.ChainID)
	data.ProvisionState = types.StringValue(activationResponse.ProvisionState)

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

	// Delete activation
	tflog.Debug(ctx, "Deleting activation", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteActivation(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete activation, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get activation from API
	tflog.Debug(ctx, "Importing activation", map[string]any{
		"id": req.ID,
	})

	activationResponse, err := r.client.Perception.GetActivation(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read activation for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(activationResponse.ID)
	data.ThoughtPathId = types.StringValue(activationResponse.ThoughtPathID)
	data.ChainId = types.StringValue(activationResponse.ChainID)
	data.ProvisionState = types.StringValue(activationResponse.ProvisionState)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
