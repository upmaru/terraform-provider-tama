// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package option

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
	"github.com/upmaru/tama-go/tools"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource { return &Resource{} }

type Resource struct{ client *tama.Client }

type ResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	ThoughtToolOutputId types.String `tfsdk:"thought_tool_output_id"`
	ActionModifierId    types.String `tfsdk:"action_modifier_id"`
	ProvisionState      types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// Name requested by user: tama_tool_output_option
	resp.TypeName = req.ProviderTypeName + "_tool_output_option"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Tool Output Option resource (binds a tool output to an action modifier)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Option identifier",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"thought_tool_output_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought tool output this option belongs to",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"action_modifier_id": schema.StringAttribute{
				MarkdownDescription: "ID of the action modifier to bind to this output",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the option",
				Computed:            true,
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := tools.CreateOptionRequest{
		Option: tools.OptionRequestData{
			ActionModifierID: data.ActionModifierId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating tool output option", map[string]any{
		"thought_tool_output_id": data.ThoughtToolOutputId.ValueString(),
		"action_modifier_id":     createReq.Option.ActionModifierID,
	})

	opt, err := r.client.Tools.CreateOption(data.ThoughtToolOutputId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tool output option, got error: %s", err))
		return
	}

	data.Id = types.StringValue(opt.ID)
	data.ThoughtToolOutputId = types.StringValue(opt.ThoughtToolOutputID)
	data.ActionModifierId = types.StringValue(opt.ActionModifierID)
	data.ProvisionState = types.StringValue(opt.ProvisionState)

	tflog.Trace(ctx, "created a tool output option resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading tool output option", map[string]any{"id": data.Id.ValueString()})
	opt, err := r.client.Tools.GetOption(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool output option, got error: %s", err))
		return
	}

	data.Id = types.StringValue(opt.ID)
	data.ThoughtToolOutputId = types.StringValue(opt.ThoughtToolOutputID)
	data.ActionModifierId = types.StringValue(opt.ActionModifierID)
	data.ProvisionState = types.StringValue(opt.ProvisionState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := tools.UpdateOptionRequest{
		Option: tools.UpdateOptionData{
			ActionModifierID: data.ActionModifierId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating tool output option", map[string]any{
		"id":                 data.Id.ValueString(),
		"action_modifier_id": updateReq.Option.ActionModifierID,
	})

	opt, err := r.client.Tools.UpdateOption(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tool output option, got error: %s", err))
		return
	}

	data.Id = types.StringValue(opt.ID)
	data.ThoughtToolOutputId = types.StringValue(opt.ThoughtToolOutputID)
	data.ActionModifierId = types.StringValue(opt.ActionModifierID)
	data.ProvisionState = types.StringValue(opt.ProvisionState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting tool output option", map[string]any{"id": data.Id.ValueString()})
	if err := r.client.Tools.DeleteOption(data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tool output option, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing tool output option", map[string]any{"id": req.ID})
	opt, err := r.client.Tools.GetOption(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool output option for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(opt.ID)
	data.ThoughtToolOutputId = types.StringValue(opt.ThoughtToolOutputID)
	data.ActionModifierId = types.StringValue(opt.ActionModifierID)
	data.ProvisionState = types.StringValue(opt.ProvisionState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}
