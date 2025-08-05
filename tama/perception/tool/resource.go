package tool

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/perception"
)

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
	ThoughtID      types.String `tfsdk:"thought_id"`
	ActionID       types.String `tfsdk:"action_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_tool"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Tool resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Tool identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this tool is attached to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action_id": schema.StringAttribute{
				MarkdownDescription: "ID of the action for this tool.",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the tool.",
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

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create tool request
	createReq := perception.CreateToolRequest{
		Tool: perception.CreateToolData{
			ActionID: data.ActionID.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating thought tool", map[string]interface{}{
		"thought_id": data.ThoughtID.ValueString(),
		"action_id": data.ActionID.ValueString(),
	})

	toolResponse, err := r.client.Perception.CreateTool(data.ThoughtID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tool, got error: %s", err))
		return
	}

	data.Id = types.StringValue(toolResponse.ID)
	data.ThoughtID = types.StringValue(toolResponse.ThoughtID)
	data.ActionID = types.StringValue(toolResponse.ActionID)
	data.ProvisionState = types.StringValue(toolResponse.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading thought tool", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	toolResponse, err := r.client.Perception.GetTool(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool, got error: %s", err))
		return
	}

	data.Id = types.StringValue(toolResponse.ID)
	data.ThoughtID = types.StringValue(toolResponse.ThoughtID)
	data.ActionID = types.StringValue(toolResponse.ActionID)
	data.ProvisionState = types.StringValue(toolResponse.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := perception.UpdateToolRequest{
		Tool: perception.UpdateToolData{
			ActionID: data.ActionID.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating thought tool", map[string]interface{}{
		"id": data.Id.ValueString(),
		"action_id": data.ActionID.ValueString(),
	})

	toolResponse, err := r.client.Perception.UpdateTool(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tool, got error: %s", err))
		return
	}

	data.Id = types.StringValue(toolResponse.ID)
	data.ThoughtID = types.StringValue(toolResponse.ThoughtID)
	data.ActionID = types.StringValue(toolResponse.ActionID)
	data.ProvisionState = types.StringValue(toolResponse.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting thought tool", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteTool(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tool, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing tool", map[string]interface{}{
		"id": req.ID,
	})

	toolResponse, err := r.client.Perception.GetTool(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(toolResponse.ID)
	data.ThoughtID = types.StringValue(toolResponse.ThoughtID)
	data.ActionID = types.StringValue(toolResponse.ActionID)
	data.ProvisionState = types.StringValue(toolResponse.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
