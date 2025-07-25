// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chain

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
	SpaceId        types.String `tfsdk:"space_id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chain"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Perception Chain resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Chain identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this chain belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the chain",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug of the chain",
				Computed:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the chain",
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

	// Create chain request
	createReq := perception.CreateChainRequest{
		Chain: perception.ChainRequestData{
			Name: data.Name.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating chain", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"name":     createReq.Chain.Name,
	})

	// Create chain
	chainResponse, err := r.client.Perception.CreateChain(data.SpaceId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create chain, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(chainResponse.ID)
	data.SpaceId = types.StringValue(chainResponse.SpaceID)
	data.Name = types.StringValue(chainResponse.Name)
	data.Slug = types.StringValue(chainResponse.Slug)
	data.ProvisionState = types.StringValue(chainResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a chain resource")

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

	// Get chain from API
	tflog.Debug(ctx, "Reading chain", map[string]any{
		"id": data.Id.ValueString(),
	})

	chainResponse, err := r.client.Perception.GetChain(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read chain, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(chainResponse.ID)
	data.SpaceId = types.StringValue(chainResponse.SpaceID)
	data.Name = types.StringValue(chainResponse.Name)
	data.Slug = types.StringValue(chainResponse.Slug)
	data.ProvisionState = types.StringValue(chainResponse.ProvisionState)

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

	// Update chain request
	updateReq := perception.UpdateChainRequest{
		Chain: perception.UpdateChainData{
			Name: data.Name.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating chain", map[string]any{
		"id":       data.Id.ValueString(),
		"space_id": data.SpaceId.ValueString(),
		"name":     updateReq.Chain.Name,
	})

	// Update chain
	chainResponse, err := r.client.Perception.UpdateChain(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update chain, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(chainResponse.ID)
	data.SpaceId = types.StringValue(chainResponse.SpaceID)
	data.Name = types.StringValue(chainResponse.Name)
	data.Slug = types.StringValue(chainResponse.Slug)
	data.ProvisionState = types.StringValue(chainResponse.ProvisionState)

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

	// Delete chain
	tflog.Debug(ctx, "Deleting chain", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Perception.DeleteChain(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete chain, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get chain from API
	tflog.Debug(ctx, "Importing chain", map[string]any{
		"id": req.ID,
	})

	chainResponse, err := r.client.Perception.GetChain(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read chain for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(chainResponse.ID)
	data.SpaceId = types.StringValue(chainResponse.SpaceID)
	data.Name = types.StringValue(chainResponse.Name)
	data.Slug = types.StringValue(chainResponse.Slug)
	data.ProvisionState = types.StringValue(chainResponse.ProvisionState)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
