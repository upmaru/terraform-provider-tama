// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package topic

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
	"github.com/upmaru/tama-go/memory"
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
	ClassId        types.String `tfsdk:"class_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_listener_topic"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Memory Listener Topic resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Topic identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"listener_id": schema.StringAttribute{
				MarkdownDescription: "ID of the listener this topic belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"class_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class this topic is associated with",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the topic",
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

	createRequest := memory.CreateTopicRequest{
		Topic: memory.TopicRequestData{
			ClassID: data.ClassId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating topic", map[string]any{
		"listener_id": data.ListenerId.ValueString(),
		"class_id":    data.ClassId.ValueString(),
	})

	topic, err := r.client.Memory.CreateTopic(data.ListenerId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create topic, got error: %s", err))
		return
	}

	data.Id = types.StringValue(topic.ID)
	data.ListenerId = types.StringValue(topic.ListenerID)
	data.ClassId = types.StringValue(topic.ClassID)
	data.ProvisionState = types.StringValue(topic.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	topic, err := r.client.Memory.GetTopic(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read topic, got error: %s", err))
		return
	}

	data.Id = types.StringValue(topic.ID)
	data.ListenerId = types.StringValue(topic.ListenerID)
	data.ClassId = types.StringValue(topic.ClassID)
	data.ProvisionState = types.StringValue(topic.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := memory.UpdateTopicRequest{
		Topic: memory.UpdateTopicData{
			ClassID: data.ClassId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating topic", map[string]any{
		"id":          data.Id.ValueString(),
		"class_id":    data.ClassId.ValueString(),
		"listener_id": data.ListenerId.ValueString(),
	})

	topic, err := r.client.Memory.UpdateTopic(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update topic, got error: %s", err))
		return
	}

	data.Id = types.StringValue(topic.ID)
	data.ListenerId = types.StringValue(topic.ListenerID)
	data.ClassId = types.StringValue(topic.ClassID)
	data.ProvisionState = types.StringValue(topic.ProvisionState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting topic", map[string]any{
		"id": data.Id.ValueString(),
	})

	if err := r.client.Memory.DeleteTopic(data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete topic, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	topic, err := r.client.Memory.GetTopic(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import topic, got error: %s", err))
		return
	}

	data := ResourceModel{
		Id:             types.StringValue(topic.ID),
		ListenerId:     types.StringValue(topic.ListenerID),
		ClassId:        types.StringValue(topic.ClassID),
		ProvisionState: types.StringValue(topic.ProvisionState),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
