// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pruning

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
	Id                    types.String `tfsdk:"id"`
	ThoughtId             types.String `tfsdk:"thought_id"`
	PreviousVersionsCount types.Int64  `tfsdk:"previous_versions_count"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_pruning"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Pruning resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Pruning identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought this pruning belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"previous_versions_count": schema.Int64Attribute{
				MarkdownDescription: "Number of previous versions to keep during pruning",
				Required:            true,
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

	if data.PreviousVersionsCount.IsUnknown() || data.PreviousVersionsCount.IsNull() {
		resp.Diagnostics.AddError("Missing Previous Versions Count", "The previous_versions_count attribute must be provided.")
		return
	}

	createReq := perception.CreatePruningRequest{
		Pruning: perception.CreatePruningData{
			PreviousVersionsCount: int(data.PreviousVersionsCount.ValueInt64()),
		},
	}

	tflog.Debug(ctx, "Creating thought pruning", map[string]any{
		"thought_id":              data.ThoughtId.ValueString(),
		"previous_versions_count": createReq.Pruning.PreviousVersionsCount,
	})

	pruningResponse, err := r.client.Perception.CreatePruning(data.ThoughtId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create thought pruning, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(pruningResponse.ID)
	data.ThoughtId = types.StringValue(pruningResponse.ThoughtID)
	data.PreviousVersionsCount = types.Int64Value(int64(pruningResponse.PreviousVersionsCount))

	tflog.Trace(ctx, "created a thought pruning resource")

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

	tflog.Debug(ctx, "Reading thought pruning", map[string]any{
		"id": data.Id.ValueString(),
	})

	pruningResponse, err := r.client.Perception.GetPruning(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read thought pruning, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.ThoughtId = types.StringValue(pruningResponse.ThoughtID)
	data.PreviousVersionsCount = types.Int64Value(int64(pruningResponse.PreviousVersionsCount))

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

	if data.PreviousVersionsCount.IsUnknown() || data.PreviousVersionsCount.IsNull() {
		resp.Diagnostics.AddError("Missing Previous Versions Count", "The previous_versions_count attribute must be provided.")
		return
	}

	previousVersionsCount := int(data.PreviousVersionsCount.ValueInt64())

	updateReq := perception.UpdatePruningRequest{
		Pruning: perception.UpdatePruningData{
			PreviousVersionsCount: &previousVersionsCount,
		},
	}

	tflog.Debug(ctx, "Updating thought pruning", map[string]any{
		"id":                      data.Id.ValueString(),
		"previous_versions_count": previousVersionsCount,
	})

	pruningResponse, err := r.client.Perception.UpdatePruning(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update thought pruning, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.ThoughtId = types.StringValue(pruningResponse.ThoughtID)
	data.PreviousVersionsCount = types.Int64Value(int64(pruningResponse.PreviousVersionsCount))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting thought pruning", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeletePruning(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete thought pruning, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pruningResponse, err := r.client.Perception.GetPruning(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import thought pruning, got error: %s", err))
		return
	}

	data := ResourceModel{
		Id:                    types.StringValue(pruningResponse.ID),
		ThoughtId:             types.StringValue(pruningResponse.ThoughtID),
		PreviousVersionsCount: types.Int64Value(int64(pruningResponse.PreviousVersionsCount)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
