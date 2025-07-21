// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package space

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
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Slug         types.String `tfsdk:"slug"`
	CurrentState types.String `tfsdk:"current_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Space resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Space identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the space",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the space (e.g., 'root', 'component')",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug identifier for the space",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the space",
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

	// Create space using the Tama client
	createRequest := neural.CreateSpaceRequest{
		Space: neural.SpaceRequestData{
			Name: data.Name.ValueString(),
			Type: data.Type.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating space", map[string]any{
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	})

	spaceResponse, err := r.client.Neural.CreateSpace(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create space, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(spaceResponse.ID)
	data.Name = types.StringValue(spaceResponse.Name)
	data.Type = types.StringValue(spaceResponse.Type)
	data.Slug = types.StringValue(spaceResponse.Slug)
	data.CurrentState = types.StringValue(spaceResponse.CurrentState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a space resource")

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

	// Get space from API
	spaceResponse, err := r.client.Neural.GetSpace(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read space, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Name = types.StringValue(spaceResponse.Name)
	data.Type = types.StringValue(spaceResponse.Type)
	data.Slug = types.StringValue(spaceResponse.Slug)
	data.CurrentState = types.StringValue(spaceResponse.CurrentState)

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

	// Update space using the Tama client
	updateRequest := neural.UpdateSpaceRequest{
		Space: neural.UpdateSpaceData{
			Name: data.Name.ValueString(),
			Type: data.Type.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating space", map[string]any{
		"id":   data.Id.ValueString(),
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	})

	spaceResponse, err := r.client.Neural.UpdateSpace(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update space, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Name = types.StringValue(spaceResponse.Name)
	data.Type = types.StringValue(spaceResponse.Type)
	data.Slug = types.StringValue(spaceResponse.Slug)
	data.CurrentState = types.StringValue(spaceResponse.CurrentState)

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

	// Delete space using the Tama client
	tflog.Debug(ctx, "Deleting space", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteSpace(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete space, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get space from API to populate state
	spaceResponse, err := r.client.Neural.GetSpace(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import space, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:           types.StringValue(spaceResponse.ID),
		Name:         types.StringValue(spaceResponse.Name),
		Type:         types.StringValue(spaceResponse.Type),
		Slug:         types.StringValue(spaceResponse.Slug),
		CurrentState: types.StringValue(spaceResponse.CurrentState),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
