// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filter

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
	ListenerId     types.String `tfsdk:"listener_id"`
	ChainId        types.String `tfsdk:"chain_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_listener_filter"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Listener Filter resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Filter identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"listener_id": schema.StringAttribute{
				MarkdownDescription: "ID of the listener this filter belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain associated with this filter",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the filter",
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

	createRequest := neural.CreateFilterRequest{
		Filter: neural.FilterRequestData{
			ChainID: data.ChainId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating filter", map[string]any{
		"listener_id": data.ListenerId.ValueString(),
		"chain_id":    data.ChainId.ValueString(),
	})

	filter, err := r.client.Neural.CreateFilter(data.ListenerId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create filter, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(filter.ID)
	data.ListenerId = types.StringValue(filter.ListenerID)
	data.ChainId = types.StringValue(filter.ChainID)
	data.ProvisionState = types.StringValue(filter.ProvisionState)

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

	filter, err := r.client.Neural.GetFilter(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read filter, got error: %s", err))
		return
	}

	data.Id = types.StringValue(filter.ID)
	data.ListenerId = types.StringValue(filter.ListenerID)
	data.ChainId = types.StringValue(filter.ChainID)
	data.ProvisionState = types.StringValue(filter.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := neural.UpdateFilterRequest{
		Filter: neural.UpdateFilterData{
			ChainID: data.ChainId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating filter", map[string]any{
		"id":          data.Id.ValueString(),
		"chain_id":    data.ChainId.ValueString(),
		"listener_id": data.ListenerId.ValueString(),
	})

	filter, err := r.client.Neural.UpdateFilter(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update filter, got error: %s", err))
		return
	}

	data.Id = types.StringValue(filter.ID)
	data.ListenerId = types.StringValue(filter.ListenerID)
	data.ChainId = types.StringValue(filter.ChainID)
	data.ProvisionState = types.StringValue(filter.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting filter", map[string]any{
		"id": data.Id.ValueString(),
	})

	if err := r.client.Neural.DeleteFilter(data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete filter, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	filter, err := r.client.Neural.GetFilter(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import filter, got error: %s", err))
		return
	}

	data := ResourceModel{
		Id:             types.StringValue(filter.ID),
		ListenerId:     types.StringValue(filter.ListenerID),
		ChainId:        types.StringValue(filter.ChainID),
		ProvisionState: types.StringValue(filter.ProvisionState),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
