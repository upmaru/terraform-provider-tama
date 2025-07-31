// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package node

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	ClassId        types.String `tfsdk:"class_id"`
	ChainId        types.String `tfsdk:"chain_id"`
	Type           types.String `tfsdk:"type"`
	On             types.String `tfsdk:"on"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Neural Node resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Node identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this node belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this node uses",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"chain_id": schema.StringAttribute{
				MarkdownDescription: "ID of the chain this node belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of node (scheduled, reactive, or explicit)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("scheduled", "reactive", "explicit"),
				},
			},
			"on": schema.StringAttribute{
				MarkdownDescription: "Event that triggers this node (default: 'processing')",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("processing"),
				Validators: []validator.String{
					stringvalidator.OneOf("processing", "processed"),
				},
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the node",
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

	// Create node using the Tama client
	createRequest := neural.CreateNodeRequest{
		Node: neural.NodeRequestData{
			Type:    data.Type.ValueString(),
			ClassID: data.ClassId.ValueString(),
			ChainID: data.ChainId.ValueString(),
		},
	}

	// Add 'on' field if it's not empty (it will have default value if not specified)
	if !data.On.IsNull() && !data.On.IsUnknown() && data.On.ValueString() != "" {
		createRequest.Node.On = data.On.ValueString()
	}

	tflog.Debug(ctx, "Creating node", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"class_id": data.ClassId.ValueString(),
		"chain_id": data.ChainId.ValueString(),
		"type":     data.Type.ValueString(),
		"on":       data.On.ValueString(),
	})

	nodeResponse, err := r.client.Neural.CreateNode(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(nodeResponse.ID)
	data.SpaceId = types.StringValue(nodeResponse.SpaceID)
	data.ClassId = types.StringValue(nodeResponse.ClassID)
	data.ChainId = types.StringValue(nodeResponse.ChainID)
	data.Type = types.StringValue(nodeResponse.Type)
	data.On = types.StringValue(nodeResponse.On)
	data.ProvisionState = types.StringValue(nodeResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a node resource")

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

	// Get node from API
	nodeResponse, err := r.client.Neural.GetNode(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Id = types.StringValue(nodeResponse.ID)
	data.SpaceId = types.StringValue(nodeResponse.SpaceID)
	data.ClassId = types.StringValue(nodeResponse.ClassID)
	data.ChainId = types.StringValue(nodeResponse.ChainID)
	data.Type = types.StringValue(nodeResponse.Type)
	data.On = types.StringValue(nodeResponse.On)
	data.ProvisionState = types.StringValue(nodeResponse.ProvisionState)

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

	// Update node using the Tama client
	updateRequest := neural.UpdateNodeRequest{
		Node: neural.UpdateNodeData{
			Type: data.Type.ValueString(),
		},
	}

	// Add 'on' field if it's not empty
	if !data.On.IsNull() && !data.On.IsUnknown() && data.On.ValueString() != "" {
		updateRequest.Node.On = data.On.ValueString()
	}

	tflog.Debug(ctx, "Updating node", map[string]any{
		"id":   data.Id.ValueString(),
		"type": data.Type.ValueString(),
		"on":   data.On.ValueString(),
	})

	nodeResponse, err := r.client.Neural.UpdateNode(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update node, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Id = types.StringValue(nodeResponse.ID)
	data.SpaceId = types.StringValue(nodeResponse.SpaceID)
	data.ClassId = types.StringValue(nodeResponse.ClassID)
	data.ChainId = types.StringValue(nodeResponse.ChainID)
	data.Type = types.StringValue(nodeResponse.Type)
	data.On = types.StringValue(nodeResponse.On)
	data.ProvisionState = types.StringValue(nodeResponse.ProvisionState)

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

	// Delete node using the Tama client
	tflog.Debug(ctx, "Deleting node", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Neural.DeleteNode(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get node from API to populate state
	nodeResponse, err := r.client.Neural.GetNode(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import node, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:             types.StringValue(nodeResponse.ID),
		SpaceId:        types.StringValue(nodeResponse.SpaceID),
		ClassId:        types.StringValue(nodeResponse.ClassID),
		ChainId:        types.StringValue(nodeResponse.ChainID),
		Type:           types.StringValue(nodeResponse.Type),
		On:             types.StringValue(nodeResponse.On),
		ProvisionState: types.StringValue(nodeResponse.ProvisionState),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
