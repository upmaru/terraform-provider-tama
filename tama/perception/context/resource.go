// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package context

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ThoughtId      types.String `tfsdk:"thought_id"`
	PromptId       types.String `tfsdk:"prompt_id"`
	Layer          types.Int64  `tfsdk:"layer"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_context"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Perception Thought Context resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Context identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this context belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt_id": schema.StringAttribute{
				MarkdownDescription: "ID of the prompt for this context",
				Required:            true,
			},
			"layer": schema.Int64Attribute{
				MarkdownDescription: "Layer number for the context",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the context",
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

	// Create context request
	createReq := perception.CreateContextRequest{
		Context: perception.ContextRequestData{
			PromptID: data.PromptId.ValueString(),
		},
	}

	// Set layer if provided, otherwise default to 0
	if !data.Layer.IsNull() && !data.Layer.IsUnknown() {
		createReq.Context.Layer = int(data.Layer.ValueInt64())
	} else {
		createReq.Context.Layer = 0
	}

	tflog.Debug(ctx, "Creating context", map[string]any{
		"thought_id": data.ThoughtId.ValueString(),
		"prompt_id":  createReq.Context.PromptID,
		"layer":      createReq.Context.Layer,
	})

	// Create context
	contextResponse, err := r.client.Perception.CreateContext(data.ThoughtId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create context, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(contextResponse.ID)
	data.ThoughtId = types.StringValue(contextResponse.ThoughtID)
	data.PromptId = types.StringValue(contextResponse.PromptID)
	data.Layer = types.Int64Value(int64(contextResponse.Layer))
	data.ProvisionState = types.StringValue(contextResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a context resource")

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

	// Get context from API
	tflog.Debug(ctx, "Reading context", map[string]any{
		"id": data.Id.ValueString(),
	})

	contextResponse, err := r.client.Perception.GetContext(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read context, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(contextResponse.ID)
	data.ThoughtId = types.StringValue(contextResponse.ThoughtID)
	data.PromptId = types.StringValue(contextResponse.PromptID)
	data.Layer = types.Int64Value(int64(contextResponse.Layer))
	data.ProvisionState = types.StringValue(contextResponse.ProvisionState)

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

	// Update context request
	updateReq := perception.UpdateContextRequest{
		Context: perception.UpdateContextData{
			PromptID: data.PromptId.ValueString(),
		},
	}

	// Set layer if provided
	if !data.Layer.IsNull() && !data.Layer.IsUnknown() {
		updateReq.Context.Layer = int(data.Layer.ValueInt64())
	}

	tflog.Debug(ctx, "Updating context", map[string]any{
		"id":        data.Id.ValueString(),
		"prompt_id": updateReq.Context.PromptID,
		"layer":     updateReq.Context.Layer,
	})

	// Update context
	contextResponse, err := r.client.Perception.UpdateContext(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update context, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(contextResponse.ID)
	data.ThoughtId = types.StringValue(contextResponse.ThoughtID)
	data.PromptId = types.StringValue(contextResponse.PromptID)
	data.Layer = types.Int64Value(int64(contextResponse.Layer))
	data.ProvisionState = types.StringValue(contextResponse.ProvisionState)

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

	// Delete context
	tflog.Debug(ctx, "Deleting context", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteContext(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete context, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get context from API
	tflog.Debug(ctx, "Importing context", map[string]any{
		"id": req.ID,
	})

	contextResponse, err := r.client.Perception.GetContext(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read context for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(contextResponse.ID)
	data.ThoughtId = types.StringValue(contextResponse.ThoughtID)
	data.PromptId = types.StringValue(contextResponse.PromptID)
	data.Layer = types.Int64Value(int64(contextResponse.Layer))
	data.ProvisionState = types.StringValue(contextResponse.ProvisionState)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
