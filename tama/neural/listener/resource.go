// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listener

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
	Endpoint       types.String `tfsdk:"endpoint"`
	Secret         types.String `tfsdk:"secret"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_listener"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Listener resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Listener identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this listener belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Destination endpoint that will receive events",
				Required:            true,
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Shared secret used to validate incoming requests",
				Required:            true,
				Sensitive:           true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the listener",
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

	createRequest := neural.CreateListenerRequest{
		Listener: neural.ListenerRequestData{
			Endpoint: data.Endpoint.ValueString(),
			Secret:   data.Secret.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating listener", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"endpoint": data.Endpoint.ValueString(),
	})

	listener, err := r.client.Neural.CreateListener(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create listener, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(listener.ID)
	data.SpaceId = types.StringValue(listener.SpaceID)
	data.Endpoint = types.StringValue(listener.Endpoint)
	data.ProvisionState = types.StringValue(listener.ProvisionState)

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

	listener, err := r.client.Neural.GetListener(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read listener, got error: %s", err))
		return
	}

	data.Id = types.StringValue(listener.ID)
	data.SpaceId = types.StringValue(listener.SpaceID)
	data.Endpoint = types.StringValue(listener.Endpoint)
	data.ProvisionState = types.StringValue(listener.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := neural.UpdateListenerRequest{
		Listener: neural.UpdateListenerData{
			Endpoint: data.Endpoint.ValueString(),
			Secret:   data.Secret.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating listener", map[string]any{
		"id":       data.Id.ValueString(),
		"endpoint": data.Endpoint.ValueString(),
	})

	listener, err := r.client.Neural.UpdateListener(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update listener, got error: %s", err))
		return
	}

	data.Id = types.StringValue(listener.ID)
	data.SpaceId = types.StringValue(listener.SpaceID)
	data.Endpoint = types.StringValue(listener.Endpoint)
	data.ProvisionState = types.StringValue(listener.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting listener", map[string]any{
		"id": data.Id.ValueString(),
	})

	if err := r.client.Neural.DeleteListener(data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete listener, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	listener, err := r.client.Neural.GetListener(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import listener, got error: %s", err))
		return
	}

	data := ResourceModel{
		Id:             types.StringValue(listener.ID),
		SpaceId:        types.StringValue(listener.SpaceID),
		Endpoint:       types.StringValue(listener.Endpoint),
		ProvisionState: types.StringValue(listener.ProvisionState),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
