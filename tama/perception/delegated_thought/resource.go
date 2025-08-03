// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package delegated_thought

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

// Resource defines the tama_delegated_thought implementation.
type Resource struct {
	client *tama.Client
}

// DelegationModel describes the delegation block data model.
type DelegationModel struct {
	TargetThoughtId types.String `tfsdk:"target_thought_id"`
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id             types.String    `tfsdk:"id"`
	ChainId        types.String    `tfsdk:"chain_id"`
	OutputClassId  types.String    `tfsdk:"output_class_id"`
	Delegation     DelegationModel `tfsdk:"delegation"`
	ProvisionState types.String    `tfsdk:"provision_state"`
	Relation       types.String    `tfsdk:"relation"`
	Index          types.Int64     `tfsdk:"index"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delegated_thought"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Perception Delegated Thought resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Delegated thought identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain this thought belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"output_class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the output class for this thought",
				Optional:            true,
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the thought",
				Computed:            true,
			},
			"relation": schema.StringAttribute{
				MarkdownDescription: "Relation type for the thought",
				Computed:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Index position of the thought in the chain",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"delegation": schema.SingleNestedBlock{
				MarkdownDescription: "Delegation configuration for the thought",
				Attributes: map[string]schema.Attribute{
					"target_thought_id": schema.StringAttribute{
						MarkdownDescription: "Target thought ID from tama_modular_thought",
						Required:            true,
					},
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan data
	if resp.Diagnostics.HasError() {
		return
	}
	delegationBlock := data.Delegation

	// Validate that target thought is a modular thought
	targetThoughtID := delegationBlock.TargetThoughtId.ValueString()
	targetThought, err := r.client.Perception.GetThought(targetThoughtID)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("Unable to validate target thought: %s", err))
		return
	}

	// Check if target thought has a module (indicating it's a modular thought)
	if targetThought.Module == nil {
		resp.Diagnostics.AddError(
			"Invalid Target Thought",
			"The target_thought_id must reference a tama_modular_thought resource. Delegated thoughts cannot reference other delegated thoughts.",
		)
		return
	}

	// Parse delegation
	delegation := perception.Delegation{
		TargetThoughtID: targetThoughtID,
	}

	createReq := perception.CreateThoughtRequest{
		Thought: perception.ThoughtRequestData{
			Delegation: &delegation,
		},
	}

	if !data.OutputClassId.IsNull() && !data.OutputClassId.IsUnknown() && data.OutputClassId.ValueString() != "" {
		createReq.Thought.OutputClassID = data.OutputClassId.ValueString()
	}

	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		index := int(data.Index.ValueInt64())
		createReq.Thought.Index = &index
	}

	tflog.Debug(ctx, "Creating delegated thought", map[string]any{
		"chain_id":      data.ChainId.ValueString(),
		"relation":      createReq.Thought.Relation,
		"delegation_id": delegation.TargetThoughtID,
	})

	thoughtResponse, err := r.client.Perception.CreateThought(data.ChainId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create delegated thought, got error: %s", err))
		return
	}

	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update delegation from response
	if thoughtResponse.Delegation != nil {
		data.Delegation.TargetThoughtId = types.StringValue(thoughtResponse.Delegation.TargetThoughtID)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	thoughtResponse, err := r.client.Perception.GetThought(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read delegated thought, got error: %s", err))
		return
	}

	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update delegation from response
	if thoughtResponse.Delegation != nil {
		data.Delegation.TargetThoughtId = types.StringValue(thoughtResponse.Delegation.TargetThoughtID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delegationBlock := data.Delegation
	delegation := perception.Delegation{
		TargetThoughtID: delegationBlock.TargetThoughtId.ValueString(),
	}

	updateReq := perception.UpdateThoughtRequest{
		Thought: perception.UpdateThoughtData{
			Relation:   data.Relation.ValueString(),
			Delegation: &delegation,
		},
	}

	if !data.OutputClassId.IsNull() && !data.OutputClassId.IsUnknown() && data.OutputClassId.ValueString() != "" {
		updateReq.Thought.OutputClassID = data.OutputClassId.ValueString()
	}

	if !data.Index.IsNull() && !data.Index.IsUnknown() {
		index := int(data.Index.ValueInt64())
		updateReq.Thought.Index = &index
	}

	thoughtResponse, err := r.client.Perception.UpdateThought(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update delegated thought, got error: %s", err))
		return
	}

	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Perception.DeleteThought(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete delegated thought, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	thoughtResponse, err := r.client.Perception.GetThought(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(thoughtResponse.ID)
	data.ChainId = types.StringValue(thoughtResponse.ChainID)
	data.OutputClassId = types.StringValue(thoughtResponse.OutputClassID)
	data.ProvisionState = types.StringValue(thoughtResponse.ProvisionState)
	data.Relation = types.StringValue(thoughtResponse.Relation)
	data.Index = types.Int64Value(int64(thoughtResponse.Index))

	// Update delegation from response
	if thoughtResponse.Delegation != nil {
		data.Delegation.TargetThoughtId = types.StringValue(thoughtResponse.Delegation.TargetThoughtID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
