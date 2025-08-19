// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	tama "github.com/upmaru/tama-go"
	"github.com/upmaru/tama-go/tools"
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
	ThoughtToolId  types.String `tfsdk:"thought_tool_id"`
	Type           types.String `tfsdk:"type"`
	ClassCorpusId  types.String `tfsdk:"class_corpus_id"`
	ProvisionState types.String `tfsdk:"provision_state"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_thought_tool_input"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tama Thought Tool Input resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Input identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"thought_tool_id": schema.StringAttribute{
				MarkdownDescription: "ID of the thought tool this input belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the input. Must be one of: path, query, header, body",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("path", "query", "header", "body"),
				},
			},
			"class_corpus_id": schema.StringAttribute{
				MarkdownDescription: "ID of the class corpus for this input",
				Required:            true,
			},
			"provision_state": schema.StringAttribute{
				MarkdownDescription: "Current state of the input",
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

	// Create input request
	createReq := tools.CreateInputRequest{
		Input: tools.InputRequestData{
			Type:          data.Type.ValueString(),
			ClassCorpusID: data.ClassCorpusId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating tool input", map[string]any{
		"thought_tool_id": data.ThoughtToolId.ValueString(),
		"type":            createReq.Input.Type,
		"class_corpus_id": createReq.Input.ClassCorpusID,
	})

	// Create input
	inputResponse, err := r.client.Tools.CreateInput(data.ThoughtToolId.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tool input, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtToolId = types.StringValue(inputResponse.ThoughtToolID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a tool input resource")

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

	// Get input from API
	tflog.Debug(ctx, "Reading tool input", map[string]any{
		"id": data.Id.ValueString(),
	})

	inputResponse, err := r.client.Tools.GetInput(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool input, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtToolId = types.StringValue(inputResponse.ThoughtToolID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

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

	// Update input request
	updateReq := tools.UpdateInputRequest{
		Input: tools.UpdateInputData{
			Type:          data.Type.ValueString(),
			ClassCorpusID: data.ClassCorpusId.ValueString(),
		},
	}

	tflog.Debug(ctx, "Updating tool input", map[string]any{
		"id":              data.Id.ValueString(),
		"type":            updateReq.Input.Type,
		"class_corpus_id": updateReq.Input.ClassCorpusID,
	})

	// Update input
	inputResponse, err := r.client.Tools.UpdateInput(data.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tool input, got error: %s", err))
		return
	}

	// Map response to resource schema
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtToolId = types.StringValue(inputResponse.ThoughtToolID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

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

	// Delete input
	tflog.Debug(ctx, "Deleting tool input", map[string]any{
		"id": data.Id.ValueString(),
	})

	err := r.client.Tools.DeleteInput(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tool input, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get input from API
	tflog.Debug(ctx, "Importing tool input", map[string]any{
		"id": req.ID,
	})

	inputResponse, err := r.client.Tools.GetInput(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tool input for import, got error: %s", err))
		return
	}

	// Map response to resource schema
	var data ResourceModel
	data.Id = types.StringValue(inputResponse.ID)
	data.ThoughtToolId = types.StringValue(inputResponse.ThoughtToolID)
	data.Type = types.StringValue(inputResponse.Type)
	data.ClassCorpusId = types.StringValue(inputResponse.ClassCorpusID)
	data.ProvisionState = types.StringValue(inputResponse.ProvisionState)

	// Save imported data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
