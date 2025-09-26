// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directive

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
	Id              types.String `tfsdk:"id"`
	ThoughtPathId   types.String `tfsdk:"thought_path_id"`
	PromptId        types.String `tfsdk:"prompt_id"`
	TargetThoughtId types.String `tfsdk:"target_thought_id"`
	ProvisionState  types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_path_directive"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Path Directive resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Directive identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_path_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought path this directive belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt_id": schema.StringAttribute{
				MarkdownDescription: "ID of the prompt for this directive",
				Required:            true,
			},
			"target_thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the target thought for this directive",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the directive",
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

	// Create directive request
	createReq := perception.CreateDirectiveRequest{
		Directive: perception.DirectiveRequestData{
			PromptID:        data.PromptId.ValueString(),
			TargetThoughtID: data.TargetThoughtId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating directive", map[string]any{
		"thought_path_id":   data.ThoughtPathId.ValueString(),
		"prompt_id":         createReq.Directive.PromptID,
		"target_thought_id": createReq.Directive.TargetThoughtID,
	})

	// Create directive
	directiveResponse, err := r.client.Perception.CreateDirective(data.ThoughtPathId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create directive, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(directiveResponse.ID)
	data.ThoughtPathId = types.StringValue(directiveResponse.ThoughtPathID)
	data.PromptId = types.StringValue(directiveResponse.PromptID)
	data.TargetThoughtId = types.StringValue(directiveResponse.TargetThoughtID)
	data.ProvisionState = types.StringValue(directiveResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a directive resource")

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

	// Get directive from API
	tflog.Debug(ctx, "Reading directive", map[string]any{
		"id": data.Id.ValueString(),
	})

	directiveResponse, err := r.client.Perception.GetDirective(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read directive, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(directiveResponse.ID)
	data.ThoughtPathId = types.StringValue(directiveResponse.ThoughtPathID)
	data.PromptId = types.StringValue(directiveResponse.PromptID)
	data.TargetThoughtId = types.StringValue(directiveResponse.TargetThoughtID)
	data.ProvisionState = types.StringValue(directiveResponse.ProvisionState)

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

	// Update directive request
	updateReq := perception.UpdateDirectiveRequest{
		Directive: perception.UpdateDirectiveData{
			PromptID:        data.PromptId.ValueString(),
			TargetThoughtID: data.TargetThoughtId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating directive", map[string]any{
		"id":                data.Id.ValueString(),
		"prompt_id":         updateReq.Directive.PromptID,
		"target_thought_id": updateReq.Directive.TargetThoughtID,
	})

	// Update directive
	directiveResponse, err := r.client.Perception.UpdateDirective(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update directive, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(directiveResponse.ID)
	data.ThoughtPathId = types.StringValue(directiveResponse.ThoughtPathID)
	data.PromptId = types.StringValue(directiveResponse.PromptID)
	data.TargetThoughtId = types.StringValue(directiveResponse.TargetThoughtID)
	data.ProvisionState = types.StringValue(directiveResponse.ProvisionState)

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

	// Delete directive
	tflog.Debug(ctx, "Deleting directive", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteDirective(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete directive, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get directive from API
	tflog.Debug(ctx, "Importing directive", map[string]any{
		"id": req.ID,
	})

	directiveResponse, err := r.client.Perception.GetDirective(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read directive for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(directiveResponse.ID)
	data.ThoughtPathId = types.StringValue(directiveResponse.ThoughtPathID)
	data.PromptId = types.StringValue(directiveResponse.PromptID)
	data.TargetThoughtId = types.StringValue(directiveResponse.TargetThoughtID)
	data.ProvisionState = types.StringValue(directiveResponse.ProvisionState)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
