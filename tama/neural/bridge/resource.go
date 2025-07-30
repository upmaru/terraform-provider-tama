// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bridge

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
	"github.com/upmaru/tama-go/neural"
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
	SpaceId        types.String `tfsdk:"space_id"`
	TargetSpaceId  types.String `tfsdk:"target_space_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_bridge"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Bridge resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Bridge identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this bridge belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the target space to bridge to",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the bridge",
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

	// Create bridge using the Tama client
	createRequest := neural.CreateBridgeRequest{
		Bridge: neural.BridgeRequestData{
			TargetSpaceID: data.TargetSpaceId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating bridge", map[string]any{
		"space_id":        data.SpaceId.ValueString(),
		"target_space_id": data.TargetSpaceId.ValueString(),
	})

	bridgeResponse, err := r.client.Neural.CreateBridge(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create bridge, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(bridgeResponse.ID)
	data.SpaceId = types.StringValue(bridgeResponse.SpaceID)
	data.TargetSpaceId = types.StringValue(bridgeResponse.TargetSpaceID)
	data.ProvisionState = types.StringValue(bridgeResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a bridge resource")

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

	// Get bridge from API
	bridgeResponse, err := r.client.Neural.GetBridge(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read bridge, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.SpaceId = types.StringValue(bridgeResponse.SpaceID)
	data.TargetSpaceId = types.StringValue(bridgeResponse.TargetSpaceID)
	data.ProvisionState = types.StringValue(bridgeResponse.ProvisionState)

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

	// Update bridge using the Tama client
	updateRequest := neural.UpdateBridgeRequest{
		Bridge: neural.UpdateBridgeData{
			TargetSpaceID: data.TargetSpaceId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating bridge", map[string]any{
		"id":              data.Id.ValueString(),
		"target_space_id": data.TargetSpaceId.ValueString(),
	})

	bridgeResponse, err := r.client.Neural.UpdateBridge(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update bridge, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.SpaceId = types.StringValue(bridgeResponse.SpaceID)
	data.TargetSpaceId = types.StringValue(bridgeResponse.TargetSpaceID)
	data.ProvisionState = types.StringValue(bridgeResponse.ProvisionState)

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

	// Delete bridge using the Tama client
	tflog.Debug(ctx, "Deleting bridge", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteBridge(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete bridge, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get bridge from API to populate state
	bridgeResponse, err := r.client.Neural.GetBridge(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import bridge, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:             types.StringValue(bridgeResponse.ID),
		SpaceId:        types.StringValue(bridgeResponse.SpaceID),
		TargetSpaceId:  types.StringValue(bridgeResponse.TargetSpaceID),
		ProvisionState: types.StringValue(bridgeResponse.ProvisionState),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
