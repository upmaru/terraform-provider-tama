// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package queue

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
	"github.com/upmaru/tama-go/system"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

// NewResource creates a new queue resource instance.
func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *tama.Client
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Role        types.String `tfsdk:"role"`
	Name        types.String `tfsdk:"name"`
	Concurrency types.Int64  `tfsdk:"concurrency"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queue"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama System Queue resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Queue identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role handled by the queue (e.g., oracle, planner)",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the queue",
				Required:            true,
			},
			"concurrency": schema.Int64Attribute{
				MarkdownDescription: "Maximum concurrent thoughts dispatched to the queue",
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

	createReq := system.CreateQueueRequest{
		Queue: system.QueueRequestData{
			Role:        data.Role.ValueString(),
			Name:        data.Name.ValueString(),
			Concurrency: int(data.Concurrency.ValueInt64()),
		},
	}

	tflog.Debug(ctx, "Creating queue", map[string]any{
		"role":        createReq.Queue.Role,
		"name":        createReq.Queue.Name,
		"concurrency": createReq.Queue.Concurrency,
	})

	queueResponse, err := r.client.System.CreateQueue(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create queue, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(queueResponse.ID)
	data.Role = types.StringValue(queueResponse.Role)
	data.Name = types.StringValue(queueResponse.Name)
	data.Concurrency = types.Int64Value(int64(queueResponse.Concurrency))

	tflog.Trace(ctx, "created a queue resource")

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

	tflog.Debug(ctx, "Reading queue", map[string]any{
		"id": data.Id.ValueString(),
	})

	queueResponse, err := r.client.System.GetQueue(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read queue, got error: %s", err))
		return
	}

	data.Id = types.StringValue(queueResponse.ID)
	data.Role = types.StringValue(queueResponse.Role)
	data.Name = types.StringValue(queueResponse.Name)
	data.Concurrency = types.Int64Value(int64(queueResponse.Concurrency))

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

	concurrency := int(data.Concurrency.ValueInt64())

	updateReq := system.UpdateQueueRequest{
		Queue: system.UpdateQueueData{
			Role:        data.Role.ValueString(),
			Name:        data.Name.ValueString(),
			Concurrency: &concurrency,
		},
	}

	tflog.Debug(ctx, "Updating queue", map[string]any{
		"id":          data.Id.ValueString(),
		"role":        updateReq.Queue.Role,
		"name":        updateReq.Queue.Name,
		"concurrency": concurrency,
	})

	queueResponse, err := r.client.System.UpdateQueue(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update queue, got error: %s", err))
		return
	}

	data.Id = types.StringValue(queueResponse.ID)
	data.Role = types.StringValue(queueResponse.Role)
	data.Name = types.StringValue(queueResponse.Name)
	data.Concurrency = types.Int64Value(int64(queueResponse.Concurrency))

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

	tflog.Debug(ctx, "Deleting queue", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.System.DeleteQueue(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete queue, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing queue", map[string]any{
		"id": req.ID,
	})

	queueResponse, err := r.client.System.GetQueue(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read queue for import, got error: %s", err))
		return
	}

	var data ResourceModel
	data.Id = types.StringValue(queueResponse.ID)
	data.Role = types.StringValue(queueResponse.Role)
	data.Name = types.StringValue(queueResponse.Name)
	data.Concurrency = types.Int64Value(int64(queueResponse.Concurrency))

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
