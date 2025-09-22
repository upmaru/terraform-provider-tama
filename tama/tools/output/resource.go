// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package output

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

// Resource defines the resource implementation.
type Resource struct{ client *tama.Client }

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id             types.String `tfsdk:"id"`
	ThoughtToolId  types.String `tfsdk:"thought_tool_id"`
	ClassCorpusId  types.String `tfsdk:"class_corpus_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_tool_output"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Tool Output resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Output identifier",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"thought_tool_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought tool this output belongs to",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"class_corpus_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class corpus for this output",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the output",
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

	createReq := tools.CreateOutputRequest{
		Output: tools.OutputRequestData{
			ClassCorpusID: data.ClassCorpusId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating tool output", map[string]any{
		"thought_tool_id": data.ThoughtToolId.ValueString(),
		"class_corpus_id": createReq.Output.ClassCorpusID,
	})

	out, err := r.client.Tools.CreateOutput(data.ThoughtToolId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tool output, got error: %s", err))
		return
	}

	data.Id = types.StringValue(out.ID)
	data.ThoughtToolId = types.StringValue(out.ThoughtToolID)
	data.ClassCorpusId = types.StringValue(out.ClassCorpusID)
	data.ProvisionState = types.StringValue(out.ProvisionState)

	tflog.Trace(ctx, "created a tool output resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading tool output", map[string]any{"id": data.Id.ValueString()})
	out, err := r.client.Tools.GetOutput(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool output, got error: %s", err))
		return
	}

	data.Id = types.StringValue(out.ID)
	data.ThoughtToolId = types.StringValue(out.ThoughtToolID)
	data.ClassCorpusId = types.StringValue(out.ClassCorpusID)
	data.ProvisionState = types.StringValue(out.ProvisionState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := tools.UpdateOutputRequest{
		Output: tools.UpdateOutputData{
			ClassCorpusID: data.ClassCorpusId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating tool output", map[string]any{
		"id":              data.Id.ValueString(),
		"class_corpus_id": updateReq.Output.ClassCorpusID,
	})

	out, err := r.client.Tools.UpdateOutput(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tool output, got error: %s", err))
		return
	}

	data.Id = types.StringValue(out.ID)
	data.ThoughtToolId = types.StringValue(out.ThoughtToolID)
	data.ClassCorpusId = types.StringValue(out.ClassCorpusID)
	data.ProvisionState = types.StringValue(out.ProvisionState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting tool output", map[string]any{"id": data.Id.ValueString()})
	if err := r.client.Tools.DeleteOutput(data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tool output, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing tool output", map[string]any{"id": req.ID})
	out, err := r.client.Tools.GetOutput(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool output for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(out.ID)
	data.ThoughtToolId = types.StringValue(out.ThoughtToolID)
	data.ClassCorpusId = types.StringValue(out.ClassCorpusID)
	data.ProvisionState = types.StringValue(out.ProvisionState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save
}
