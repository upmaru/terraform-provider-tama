// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package prompt

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
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	SpaceId      types.String `tfsdk:"space_id"`
	Slug         types.String `tfsdk:"slug"`
	Content      types.String `tfsdk:"content"`
	Role         types.String `tfsdk:"role"`
	CurrentState types.String `tfsdk:"current_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Memory Prompt resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Prompt identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the prompt",
				Required:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "ID of the space this prompt belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Slug for the prompt",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Content of the prompt",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role associated with the prompt (system or user)",
				Required:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the prompt",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	// Create prompt using the Tama client
	createRequest := memory.CreatePromptRequest{
		Prompt: memory.PromptRequestData{
			Name:    data.Name.ValueString(),
			Content: data.Content.ValueString(),
			Role:    data.Role.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating prompt", map[string]any{
		"space_id": data.SpaceId.ValueString(),
		"name":     data.Name.ValueString(),
		"content":  data.Content.ValueString(),
		"role":     data.Role.ValueString(),
	})

	promptResponse, err := r.client.Memory.CreatePrompt(data.SpaceId.ValueString(), createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create prompt, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(promptResponse.ID)
	data.Name = types.StringValue(promptResponse.Name)
	data.Slug = types.StringValue(promptResponse.Slug)
	data.Content = types.StringValue(promptResponse.Content)
	data.Role = types.StringValue(promptResponse.Role)
	data.CurrentState = types.StringValue(promptResponse.CurrentState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a prompt resource")

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

	// Get prompt from API
	promptResponse, err := r.client.Memory.GetPrompt(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prompt, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Name = types.StringValue(promptResponse.Name)
	data.Slug = types.StringValue(promptResponse.Slug)
	data.Content = types.StringValue(promptResponse.Content)
	data.Role = types.StringValue(promptResponse.Role)
	data.CurrentState = types.StringValue(promptResponse.CurrentState)

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

	// Update prompt using the Tama client
	updateRequest := memory.UpdatePromptRequest{
		Prompt: memory.UpdatePromptData{
			Name:    data.Name.ValueString(),
			Content: data.Content.ValueString(),
			Role:    data.Role.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating prompt", map[string]any{
		"id":      data.Id.ValueString(),
		"name":    data.Name.ValueString(),
		"content": data.Content.ValueString(),
		"role":    data.Role.ValueString(),
	})

	promptResponse, err := r.client.Memory.UpdatePrompt(data.Id.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update prompt, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.Name = types.StringValue(promptResponse.Name)
	data.Slug = types.StringValue(promptResponse.Slug)
	data.Content = types.StringValue(promptResponse.Content)
	data.Role = types.StringValue(promptResponse.Role)
	data.CurrentState = types.StringValue(promptResponse.CurrentState)

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

	// Delete prompt using the Tama client
	tflog.Debug(ctx, "Deleting prompt", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Memory.DeletePrompt(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete prompt, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get prompt from API to populate state
	promptResponse, err := r.client.Memory.GetPrompt(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import prompt, got error: %s", err))
		return
	}

	// Create model from API response
	data := ResourceModel{
		Id:           types.StringValue(promptResponse.ID),
		Name:         types.StringValue(promptResponse.Name),
		SpaceId:      types.StringValue(promptResponse.SpaceID),
		Slug:         types.StringValue(promptResponse.Slug),
		Content:      types.StringValue(promptResponse.Content),
		Role:         types.StringValue(promptResponse.Role),
		CurrentState: types.StringValue(promptResponse.CurrentState),
	}

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
